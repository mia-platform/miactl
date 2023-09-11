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

package environments

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listEnvironmentsEndpointTemplate = "api/backend/projects/%s"
	getClusterEndpointTemplate       = "api/tenants/%s/clusters/%s"
)

func EnvironmentCmd(o *clioptions.CLIOptions) *cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:   "environment",
		Short: "Manage resources deployed with Mia-Platform Console",
		Long: `Manage resources deployed with Mia-Platform Console.

Once a project from Mia-Platform Console is deployed at least once, you can manage and monitor
the resources generated, like Pods, Cronjobs and logs.
`,
	}

	// add sub commands
	runtimeCmd.AddCommand(
		listEnvironmentsCmd(o),
	)

	return runtimeCmd
}

func listEnvironmentsCmd(o *clioptions.CLIOptions) *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List all environments for a given project id",
		Long:  "List all environments for a given project id",
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printEnvironments(client, restConfig.CompanyID, restConfig.ProjectID)
		},
	}

	return listCommand
}

func printEnvironments(client *client.APIClient, companyID, projectID string) error {
	switch {
	case len(companyID) == 0:
		return fmt.Errorf("missing company id, please set one with the flag or context")
	case len(projectID) == 0:
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEnvironmentsEndpointTemplate, projectID)).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	var project resources.Project
	if err := resp.ParseResponse(&project); err != nil {
		return fmt.Errorf("error parsing response body: %w", err)
	}

	if project.CompanyID != companyID {
		return fmt.Errorf("no project found with this id in the current company")
	}

	environments := project.Environments
	if len(environments) == 0 {
		fmt.Printf("No environment found for %s project\n", project.Name)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Name", "Enironment ID", "Production", "Cluster", "Kubernetes Namespace"})

	clustersCache := make(map[string]string, 0)
	for _, env := range environments {
		clusterID := env.Cluster.ID
		clusterName, found := clustersCache[clusterID]
		if !found {
			name, err := clusterNameForID(client, companyID, clusterID)
			if err != nil {
				return err
			}
			clustersCache[clusterID] = name
			clusterName = name
		}

		table.Append([]string{
			env.DisplayName,
			env.EnvID,
			strconv.FormatBool(env.IsProduction),
			clusterName,
			env.Cluster.Namespace},
		)
	}

	table.Render()
	return nil
}

func clusterNameForID(client *client.APIClient, companyID, clusterID string) (string, error) {
	if len(clusterID) == 0 {
		return "", nil
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(getClusterEndpointTemplate, companyID, clusterID)).
		Do(context.Background())

	if err != nil {
		return "", fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return "", err
	}

	var cluster resources.Cluster
	if err := resp.ParseResponse(&cluster); err != nil {
		return "", fmt.Errorf("error parsing response body: %w", err)
	}

	return cluster.DisplayName, nil
}
