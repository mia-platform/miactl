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

package get

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/resources"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	oktaProvider = "okta"
	projectsURI  = "/api/backend/projects/"
)

var (
	validArgs = []string{
		"project", "projects",
		"deployment", "deployments",
	}
)

// NewGetCmd func creates a new command
func NewGetCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:       "get",
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "projects", "project":
			case "deployment", "deployments":
				return cmd.MarkFlagRequired("project")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			switch resource {
			case "projects", "project":
				mc, err := httphandler.ConfigureDefaultMiaClient(options, projectsURI)
				if err != nil {
					return err
				}
				if err := getProjects(mc, options); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unexpected argument: %s", resource)
			}
			return nil
		},
	}
}

// getProjects retrieves the projects with the company ID of the current context
func getProjects(mc *httphandler.MiaClient, opts *clioptions.CLIOptions) error {

	// execute the request
	resp, err := mc.GetSession().Get().ExecuteRequest()
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer resp.Body.Close()

	var projects []resources.Project
	currentContext := mc.GetSession().GetContext()

	if resp.StatusCode == http.StatusOK {
		companyID, err := context.GetContextCompanyID(currentContext)
		if err != nil {
			return fmt.Errorf("error retrieving company ID for context %s: %w", currentContext, err)
		}
		if err := httphandler.ParseResponseBody(currentContext, resp.Body, projects); err != nil {
			return fmt.Errorf("error parsing response body: %w", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Configuration Git Path", "Project ID"})
		for _, project := range projects {
			if project.TenantID == companyID {
				table.Append([]string{project.Name, project.ConfigurationGitPath, project.ProjectID})
			}
		}
		table.Render()
	} else {
		return fmt.Errorf("request failed with status code: %s", resp.Status)
	}

	return nil

}
