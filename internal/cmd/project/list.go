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
	"fmt"
	"net/http"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	oktaProvider = "okta"
	projectsURI  = "/api/backend/projects/"
)

// NewListProjectsCmd func creates a new command
func NewListProjectsCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list mia projects in the current context",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			currentContext, err := context.GetCurrentContext()
			if err != nil {
				return err
			}
			return context.SetContextValues(cmd, currentContext)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mc, err := httphandler.ConfigureDefaultMiaClient(options, projectsURI)
			if err != nil {
				return err
			}
			return listProjects(mc, options)
		},
	}
}

// listProjects retrieves the projects with the company ID of the current context
func listProjects(mc *httphandler.MiaClient, opts *clioptions.CLIOptions) error {
	// execute the request
	resp, err := mc.GetSession().Get().ExecuteRequest()
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer resp.Body.Close()

	var projects []resources.Project
	currentContext := mc.GetSession().GetContext()

	if resp.StatusCode == http.StatusOK {
		companyID := opts.CompanyID
		if companyID == "" {
			return fmt.Errorf("please set a company ID for context %s", currentContext)
		}
		if err := httphandler.ParseResponseBody(currentContext, resp.Body, &projects); err != nil {
			return fmt.Errorf("error parsing response body: %w", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeader([]string{"Name", "Project ID", "Configuration Git Path"})
		for _, project := range projects {
			if project.TenantID == companyID {
				table.Append([]string{project.Name, project.ProjectID, project.ConfigurationGitPath})
			}
		}
		table.Render()
	} else {
		return fmt.Errorf("request failed with status code: %s", resp.Status)
	}

	return nil
}
