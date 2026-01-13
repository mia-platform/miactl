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

package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	deploymentsLatestEndpointTemplate = "/api/deploy/projects/%s/deployment/"
)

func NewDeploymentsCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployments",
		Short: "Manage deployments of the project",
	}

	cmd.AddCommand(
		newLatestDeploymentCmd(options),
	)

	return cmd
}

func newLatestDeploymentCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Get the latest deployment for the project",
		Long:  "Get the latest deployment for the project in the specified environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLatestDeployment(cmd.Context(), options)
		},
	}

	flags := cmd.Flags()
	options.AddDeployLatestFlags(flags)

	return cmd
}

func runLatestDeployment(ctx context.Context, options *clioptions.CLIOptions) error {
	restConfig, err := options.ToRESTConfig()
	if err != nil {
		return err
	}

	projectID := restConfig.ProjectID
	if len(projectID) == 0 {
		return errors.New("projectId is required")
	}

	client, err := client.APIClientForConfig(restConfig)
	if err != nil {
		return err
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(deploymentsLatestEndpointTemplate, projectID)).
		SetParam("page", "1").
		SetParam("per_page", "1").
		SetParam("scope", "success").
		SetParam("environment", options.Environment).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	var deployments []resources.DeploymentHistory
	if err := resp.ParseResponse(&deployments); err != nil {
		return fmt.Errorf("cannot parse server response: %w", err)
	}

	if len(deployments) == 0 {
		fmt.Println("No successful deployments found")
		return nil
	}

	latest := deployments[0]
	fmt.Printf("Latest deployment for environment %s:\n", latest.Environment)
	fmt.Printf("ID: %s\n", latest.ID)
	fmt.Printf("Ref: %s\n", latest.Ref)
	fmt.Printf("Status: %s\n", latest.Status)
	fmt.Printf("Finished At: %s\n", latest.FinishedAt)

	return nil
}
