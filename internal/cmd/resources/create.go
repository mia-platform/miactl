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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	createJobTemplate = "/api/projects/%s/environments/%s/jobs/"
)

func CreateCommand(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Mia-Platform Console runtime resources",
		Long:  "Create Mia-Platform Console runtime resources.",
	}

	// add cmd flags
	options.AddEnvironmentFlags(cmd.Flags())

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
			return createJob(cmd.Context(), client, restConfig.ProjectID, restConfig.Environment, options.FromCronJob)
		},
	}

	// add cmd flags
	options.AddCreateJobFlags(cmd.Flags())
	if err := cmd.MarkFlagRequired("from"); err != nil {
		// programming error, panic and broke everything
		panic(err)
	}

	return cmd
}

func createJob(ctx context.Context, client *client.APIClient, projectID, environment, cronjobName string) error {
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
	return nil
}
