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

package deploy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	deployProjectEndpointTemplate  = "/api/deploy/projects/%s/trigger/pipeline/"
	pipelineStatusEndpointTemplate = "/api/deploy/projects/%s/pipelines/%d/status/"
)

func triggerCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger ENVIRONMENT",
		Short: "Deploy the target environment.",
		Long: `Trigger the deploy of the target environment in the selected project.

The deploy will be performed by the pipeline setup in project, the command will then keep
listening on updates of the status for keep the user informed on the updates. The command
will exit with error if the pipeline will not end with a success.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			environmentName := args[0]
			return runDeployTrigger(cmd.Context(), environmentName, options)
		},
	}

	deployTriggerOptions(cmd, options)

	return cmd
}

func deployTriggerOptions(cmd *cobra.Command, options *clioptions.CLIOptions) {
	// set flags
	flags := cmd.Flags()
	options.AddConnectionFlags(flags)
	options.AddContextFlags(flags)
	options.AddCompanyFlags(flags)
	options.AddProjectFlags(flags)
	options.AddDeployFlags(flags)
	if err := cmd.MarkFlagRequired("revision"); err != nil {
		// if there is an error something very wrong is happening, panic
		panic(err)
	}
}

func runDeployTrigger(ctx context.Context, environmentName string, options *clioptions.CLIOptions) error {
	if len(options.Revision) == 0 {
		return errors.New("a valid revision is required to start a deploy")
	}

	restConfig, err := options.ToRESTConfig()
	if err != nil {
		return err
	}

	projectID := restConfig.ProjectID
	if len(projectID) == 0 {
		return errors.New("projectId is required to start a deploy")
	}

	client, err := client.APIClientForConfig(restConfig)
	if err != nil {
		return err
	}

	resp, err := triggerPipeline(ctx, client, environmentName, projectID, options)
	if err != nil {
		return fmt.Errorf("error executing the deploy request: %w", err)
	}
	fmt.Printf("Deploying project %s in the environment '%s'\n", projectID, environmentName)

	status, err := waitStatus(ctx, client, projectID, resp.ID, environmentName)
	if err != nil {
		return fmt.Errorf("error retrieving the pipeline status: %w", err)
	}

	if status == "failed" {
		return errors.New("pipeline failed")
	}

	fmt.Printf("Pipeline ended with %s\n", status)
	return nil
}

func triggerPipeline(ctx context.Context, client *client.APIClient, environmentName, projectID string, options *clioptions.CLIOptions) (*resources.DeployProject, error) {
	request := resources.DeployProjectRequest{
		Environment: environmentName,
		Revision:    options.Revision,
		Type:        options.DeployType,
		ForceDeploy: options.NoSemVer,
	}

	if options.DeployType == "deploy_all" {
		request.ForceDeploy = true
	}

	requestBody, err := resources.EncodeResourceToJSON(request)
	if err != nil {
		return nil, fmt.Errorf("error mashalling body: %w", err)
	}

	resp, err := client.
		Post().
		APIPath(fmt.Sprintf(deployProjectEndpointTemplate, projectID)).
		Body(requestBody).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	body := new(resources.DeployProject)
	err = resp.ParseResponse(body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// one and a half second is the time we wait between calls to the status endpoint.
// Declared here to override it during tests
var sleepDuration = (1 * time.Second) + (500 * time.Millisecond)

func waitStatus(ctx context.Context, client *client.APIClient, projectID string, deployID int, environmentName string) (string, error) {
	var outStatus *resources.PipelineStatus
	for {
		time.Sleep(sleepDuration)
		resp, err := client.
			Get().
			APIPath(fmt.Sprintf(pipelineStatusEndpointTemplate, projectID, deployID)).
			SetParam("environment", environmentName).
			Do(ctx)

		if err != nil {
			return "", err
		}
		if err := resp.Error(); err != nil {
			return "", err
		}

		status := new(resources.PipelineStatus)
		if err := resp.ParseResponse(status); err != nil {
			return "", err
		}

		if status.Status != "running" && status.Status != "pending" {
			outStatus = status
			break
		}

		fmt.Printf("The pipeline is %s..\n", status.Status)
	}

	return outStatus.Status, nil
}
