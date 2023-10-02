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

package deployments

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listEndpointTemplate = "/api/projects/%s/environments/%s/deployments/describe/"
)

func Command(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "Manage Mia-Platform Console project runtime deployment resources",
		Long: `Manage Mia-Platform Console project runtime deployment resources.

A project on Mia-Platform Console once deployed can have one or more deployment resources associcated with one or more
of its environments.
`,
	}

	// add sub commands
	cmd.AddCommand(
		listCmd(o),
	)

	return cmd
}

func listCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list ENVIRONMENT",
		Short: "List all deployments for a project in an environment",
		Long:  "List all deployments for a project in an environment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printDeploymentsList(client, restConfig.ProjectID, args[0])
		},
	}

	return cmd
}

func printDeploymentsList(client *client.APIClient, projectID, environment string) error {
	if projectID == "" {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}
	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEndpointTemplate, projectID, environment)).
		Do(context.Background())

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	deployments := make([]resources.Deployment, 0)
	err = resp.ParseResponse(&deployments)
	if err != nil {
		return err
	}

	if len(deployments) == 0 {
		fmt.Printf("No deployments found for %s environment\n", environment)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Name", "Ready", "Up-to-Date", "Available", "Age"})

	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		table.Append(rowForDeployment(deployment))
	}

	table.Render()
	return nil
}

func rowForDeployment(deployment resources.Deployment) []string {
	return []string{
		deployment.Name,
		fmt.Sprintf("%d/%d", deployment.Ready, deployment.Available),
		fmt.Sprint(deployment.Replicas),
		fmt.Sprint(deployment.Available),
		util.HumanDuration(time.Since(deployment.Age)),
	}
}
