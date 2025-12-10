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
	if err := validateCreateJobParams(projectID, environment); err != nil {
		return err
	}

	jobName, err := triggerJobCreation(ctx, client, projectID, environment, cronjobName)
	if err != nil {
		return err
	}

	fmt.Printf("Job %s created successfully!\n", jobName)

	if !waitJobCompletion {
		return nil
	}

	return waitForJobCompletionWithInterval(ctx, client, projectID, environment, jobName, waitJobTimeoutSeconds, 10*time.Second)
}

func validateCreateJobParams(projectID, environment string) error {
	if projectID == "" {
		return errors.New("missing project id, please set one with the flag or context")
	}
	if environment == "" {
		return errors.New("missing environment, please set one with the flag or context")
	}
	return nil
}

func triggerJobCreation(ctx context.Context, client *client.APIClient, projectID, environment, cronjobName string) (string, error) {
	requestBody := &resources.CreateJobRequest{
		From:         "cronjob",
		ResourceName: cronjobName,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	response, err := client.
		Post().
		APIPath(fmt.Sprintf(createJobTemplate, projectID, environment)).
		Body(bodyBytes).
		Do(ctx)
	if err != nil {
		return "", err
	}

	if err := response.Error(); err != nil {
		return "", err
	}

	var createResponse resources.CreateJob
	if err := response.ParseResponse(&createResponse); err != nil {
		return "", err
	}

	return createResponse.JobName, nil
}

func waitForJobCompletion(ctx context.Context, client *client.APIClient, projectID, environment, jobName string, timeoutSeconds int) error {
	return waitForJobCompletionWithInterval(ctx, client, projectID, environment, jobName, timeoutSeconds, 10*time.Second)
}

func waitForJobCompletionWithInterval(ctx context.Context, client *client.APIClient, projectID, environment, jobName string, timeoutSeconds int, tickerInterval time.Duration) error {
	fmt.Printf("Waiting for job %s to complete (timeout: %ds)...\n", jobName, timeoutSeconds)

	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	const maxRetries = 3
	retries := 0

	for {
		select {
		case <-timeout:
			return fmt.Errorf("job %s did not complete within %d seconds", jobName, timeoutSeconds)
		case <-ticker.C:
			completed, err := checkJobStatus(ctx, client, projectID, environment, jobName)
			if err != nil {
				retries++
				if retries >= maxRetries {
					return fmt.Errorf("max retries reached while checking job status: %w", err)
				}
				fmt.Printf("Error checking job status (retry %d/%d): %v\n", retries, maxRetries, err)
				continue
			}

			retries = 0
			if completed {
				fmt.Printf("Job %s completed successfully!\n", jobName)
				return nil
			}
		}
	}
}

func checkJobStatus(ctx context.Context, client *client.APIClient, projectID, environment, jobName string) (bool, error) {
	job, err := fetchJobStatus(ctx, client, projectID, environment, jobName)
	if err != nil {
		return false, err
	}

	pods, err := fetchJobPods(ctx, client, projectID, environment, jobName)
	if err != nil {
		fmt.Printf("Warning: could not retrieve pods for job %s: %v\n", jobName, err)
		pods = []*resources.Pod{}
	}

	printJobStatus(job, pods)

	return job.Succeeded > 0, nil
}

func fetchJobStatus(ctx context.Context, client *client.APIClient, projectID, environment, jobName string) (*resources.Job, error) {
	response, err := client.
		Get().
		APIPath(fmt.Sprintf(describeJobsTemplate, projectID, environment)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve jobs: %w", err)
	}

	var jobs []resources.Job
	if err := response.ParseResponse(&jobs); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	job := getCreatedJobFromDescribe(&jobs, jobName)
	if job == nil {
		return nil, fmt.Errorf("job %s not found", jobName)
	}

	return job, nil
}

func fetchJobPods(ctx context.Context, client *client.APIClient, projectID, environment, jobName string) ([]*resources.Pod, error) {
	response, err := client.
		Get().
		APIPath(fmt.Sprintf(describePodsTemplate, projectID, environment)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pods: %w", err)
	}

	var pods []resources.Pod
	if err := response.ParseResponse(&pods); err != nil {
		return nil, fmt.Errorf("failed to parse pods response: %w", err)
	}

	return getJobPodsFromDescribe(&pods, jobName), nil
}

func printJobStatus(job *resources.Job, pods []*resources.Pod) {
	fmt.Printf("Job Status - Active: %d | Pods: %d | Succeeded: %d | Failed: %d\n",
		job.Active, len(pods), job.Succeeded, job.Failed)

	for _, pod := range pods {
		fmt.Printf("  └─ Pod: %s | Phase: %s | Status: %s | Age: %s\n",
			pod.Name, pod.Phase, pod.Status, pod.Age.Format(time.RFC3339))
	}
}

func getCreatedJobFromDescribe(jobsDescribe *[]resources.Job, jobName string) *resources.Job {
	for i := range *jobsDescribe {
		if (*jobsDescribe)[i].Name == jobName {
			return &(*jobsDescribe)[i]
		}
	}
	return nil
}

func getJobPodsFromDescribe(podsDescribe *[]resources.Pod, jobName string) []*resources.Pod {
	jobPods := make([]*resources.Pod, 0)

	for i := range *podsDescribe {
		if (*podsDescribe)[i].Labels["job-name"] == jobName {
			jobPods = append(jobPods, &(*podsDescribe)[i])
		}
	}

	sort.Slice(jobPods, func(i, j int) bool {
		return jobPods[i].Age.Before(jobPods[j].Age)
	})

	return jobPods
}
