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

package project

import (
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	listProjectsEndpoint = "/api/backend/projects/"
)

// ListCmd return a cobra command for listing projects
func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	prjListCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects for the current user",
		Long: `List projects for the current user in the selected company.

The company can be set via the dedicated flag, or it will be inferred from
the current context. If no company can be selected the command will return
an error.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return listProjects(cmd.Context(), client, restConfig.CompanyID, options.Printer())
		},
	}

	return prjListCmd
}

// listProjects retrieves the projects with the company ID of the current context
func listProjects(ctx context.Context, client *client.APIClient, companyID string, p printer.IPrinter) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}

	// execute the request
	resp, err := client.
		Get().
		SetParam("tenantIds", companyID).
		APIPath(listProjectsEndpoint).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	projects := make([]*resources.Project, 0)
	if err := resp.ParseResponse(&projects); err != nil {
		return fmt.Errorf("error parsing response body: %w", err)
	}

	p.Keys("Name", "Project ID", "Configuration Git Path")
	for _, project := range projects {
		if project.CompanyID == companyID {
			p.Record(project.Name, project.ID, project.ConfigurationGitPath)
		}
	}

	p.Print()
	return nil
}
