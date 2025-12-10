// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	createJobTemplate    = "/api/projects/%s/environments/%s/jobs/"
	describeJobsTemplate = "/api/projects/%s/environments/%s/jobs/describe/"
	describePodsTemplate = "/api/projects/%s/environments/%s/pods/describe/"
)

func CreateCommand(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Mia-Platform Console runtime resources",
		Long:  "Create Mia-Platform Console runtime resources.",
	}

	// add sub commands
	cmd.AddCommand(
		jobCommand(options),
	)

	return cmd
}

func jobCommand(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Create a job from a cronjob in the selected environment and project",
		Long:  "Create a job from a cronjob in the selected environment and project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			err = createJob(cmd.Context(), client, restConfig.ProjectID, restConfig.Environment, options.FromCronJob, options.WaitJobCompletion, options.WaitJobTimeoutSeconds)
			if err != nil {
				// Silence usage for runtime errors
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	// add cmd flags
	flags := cmd.Flags()
	options.AddCreateJobFlags(flags)
	options.AddEnvironmentFlags(flags)
	if err := cmd.MarkFlagRequired("from"); err != nil {
		// programming error, panic and broke everything
		panic(err)
	}

	return cmd
}

func createJob(ctx context.Context, client *client.APIClient, projectID, environment, cronjobName string, waitJobCompletion bool, waitJobTimeoutSeconds int) error {
	if projectID == "" {
		return errors.New("missing project id, please set one with the flag or context")
	}

	if environment == "" {
		return errors.New("missing environment, please set one with the flag or context")
	}

	requestBody := &resources.CreateJobRequest{
		From:         "cronjob",
		ResourceName: cronjobName,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	response, err := client.
		Post().
		APIPath(fmt.Sprintf(createJobTemplate, projectID, environment)).
		Body(bodyBytes).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := response.Error(); err != nil {
		return err
	}

	var createResponse resources.CreateJob
	if err := response.ParseResponse(&createResponse); err != nil {
		return err
	}

	fmt.Printf("Job %s create successfully!\n", createResponse.JobName)
	if !waitJobCompletion {
		return nil
	}
	fmt.Printf("Waiting for job %s to complete...\n", createResponse.JobName)
	now := time.Now()
	jobCompleted := false
	retries := 0
	for !jobCompleted && time.Since(now) < time.Duration(waitJobTimeoutSeconds)*time.Second && retries < 3 {
		time.Sleep(10 * time.Second)
		listJobsResponse, err := client.
			Get().
			APIPath(fmt.Sprintf(describeJobsTemplate, projectID, environment)).
			Do(ctx)
		if err != nil {
			fmt.Printf("Error retrieving jobs from Console\n")
			retries++
			continue
		}

		var jobsDescribe []resources.Job
		if err := listJobsResponse.ParseResponse(&jobsDescribe); err != nil {
			fmt.Printf("Error parsing jobs describe response: %v\n", err)
		}

		currentJob := getCreatedJobFromDescribe(&jobsDescribe, createResponse.JobName)
		if currentJob == nil {
			fmt.Printf("Error - Created job %s not found in Console\n", createResponse.JobName)
			retries++
			continue
		}

		listPodsResponse, err := client.
			Get().
			APIPath(fmt.Sprintf(describePodsTemplate, projectID, environment)).
			Do(ctx)
		if err != nil {
			fmt.Printf("Error retrieving pods from Console\n")
		}

		var podsDescribe []resources.Pod
		if err := listPodsResponse.ParseResponse(&podsDescribe); err != nil {
			fmt.Printf("Error parsing pods describe response: %v\n", err)
		}

		currentJobPods := getJobPodsFromDescribe(&podsDescribe, createResponse.JobName)
		if len(currentJobPods) == 0 {
			fmt.Printf("Error - No pods found for job %s in Console\n", createResponse.JobName)
		}

		fmt.Printf("Job Active: %d | Pods: %d | Succeeded: %d | Failed: %d\n", currentJob.Active, len(currentJobPods), currentJob.Succeeded, currentJob.Failed)
		for _, pod := range currentJobPods {
			fmt.Printf(" - Pod %s | Phase: %s | Status: %s | Age: %s\n", pod.Name, pod.Phase, pod.Status, pod.Age.String())
		}

		if currentJob.Succeeded > 0 {
			jobCompleted = true
			fmt.Printf("Job %s completed successfully!\n", createResponse.JobName)
			break
		}
	}
	if !jobCompleted {
		return fmt.Errorf("job %s did not complete within %d seconds", createResponse.JobName, waitJobTimeoutSeconds)
	}
	return nil
}

func getCreatedJobFromDescribe(jobsDescribe *[]resources.Job, jobName string) *resources.Job {
	for _, job := range *jobsDescribe {
		if job.Name == jobName {
			return &job
		}
	}
	return nil
}

func getJobPodsFromDescribe(podsDescribe *[]resources.Pod, jobName string) []*resources.Pod {
	jobPods := []*resources.Pod{}
	for _, pod := range *podsDescribe {
		if pod.Labels["job-name"] == jobName {
			jobPods = append(jobPods, &pod)
		}
	}

	sort.Slice(jobPods, func(i, j int) bool {
		return jobPods[i].Age.Before(jobPods[j].Age)
	})

	return jobPods
}
