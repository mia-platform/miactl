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
	"fmt"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/deploy"
	"github.com/spf13/cobra"
)

const (
	deployStatusTriggerEndpointTemplate = "/api/deploy/webhooks/projects/%s/pipelines/triggers/%s/status/"
	deployStatusErrorRequiredTemplate   = "%s is required to update the deploy trigger status"
)

var allowedArgs = []string{"success", "failed", "canceled", "skipped"}

func newStatusAddCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status" + " [" + strings.Join(allowedArgs, "|") + "]",
		Short: "Add status to deploy history record.",
		Long: `This command is used to add a status to a deploy history record.

The status can be updated only once, using the trigger ID provided in the 'deploy trigger' command
to the pipeline.

At the moment, the only deploy trigger which creates a trigger ID is the integration with the Jenkins provider.`,
		ValidArgs: allowedArgs,
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			cobra.OnlyValidArgs,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddDeployStatus(cmd.Context(), options, args[0])
		},
	}

	flags := cmd.Flags()
	options.AddConnectionFlags(flags)
	options.AddContextFlags(flags)
	options.AddCompanyFlags(flags)
	options.AddProjectFlags(flags)
	options.AddDeployAddStatusFlags(flags)
	if err := cmd.MarkFlagRequired("trigger-id"); err != nil {
		// if there is an error something very wrong is happening, panic
		panic(err)
	}

	return cmd
}

func runAddDeployStatus(ctx context.Context, options *clioptions.CLIOptions, status string) error {
	restConfig, err := options.ToRESTConfig()
	if err != nil {
		return err
	}

	projectID := restConfig.ProjectID
	if len(projectID) == 0 {
		return fmt.Errorf(deployStatusErrorRequiredTemplate, "projectId")
	}

	client, err := client.APIClientForConfig(restConfig)
	if err != nil {
		return err
	}

	triggerID := options.TriggerID
	if len(triggerID) == 0 {
		return fmt.Errorf(deployStatusErrorRequiredTemplate, "triggerId")
	}

	requestBody := deploy.AddPipelineStatusRequest{
		Status: status,
	}
	payload, err := resources.EncodeResourceToJSON(requestBody)
	if err != nil {
		return err
	}

	resp, err := client.Post().
		APIPath(fmt.Sprintf(deployStatusTriggerEndpointTemplate, projectID, triggerID)).
		Body(payload).
		Do(ctx)
	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("Deploy status updated for pipeline with triggerId %s to %s\n", triggerID, status)

	return nil
}
