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

package cmd

import (
	"strconv"
	"time"

	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/renderer"
	"github.com/mia-platform/miactl/old/sdk"
	"github.com/mia-platform/miactl/old/sdk/deploy"
	"github.com/spf13/cobra"
)

var validArgs = []string{
	"project", "projects",
	"deployment", "deployments",
}

// NewGetCmd func creates a new command
func newGetCmd() *cobra.Command {
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
			f, err := factory.FromContext(cmd.Context(), sdk.Options{})
			if err != nil {
				return err
			}

			resource := args[0]

			switch resource {
			case "projects", "project":
				getProjects(f)
			case "deployment", "deployments":
				getDeploysForProject(f)
			}
			return nil
		},
	}
}

func getProjects(f *factory.Factory) {
	projects, err := f.MiaClient.Projects.Get()
	if err != nil {
		f.Renderer.Error(err).Render()
		return
	}

	headers := []string{"#", "Name", "Configuration Git Path", "Project id"}
	table := f.Renderer.Table(headers)
	for i, project := range projects {
		table.Append([]string{
			strconv.Itoa(i + 1),
			project.Name,
			project.ConfigurationGitPath,
			project.ProjectID,
		})
	}
	table.Render()
}

func getDeploysForProject(f *factory.Factory) {
	query := deploy.HistoryQuery{
		ProjectID: projectID,
	}

	history, err := f.MiaClient.Deploy.GetHistory(query)
	if err != nil {
		f.Renderer.Error(err).Render()
		return
	}

	headers := []string{"#", "Status", "Deploy Type", "Environment", "Deploy Branch/Tag", "Made By", "Duration", "Finished At", "View Log"}
	table := f.Renderer.Table(headers)
	for _, deploy := range history {
		table.Append([]string{
			strconv.Itoa(deploy.ID),
			deploy.Status,
			deploy.DeployType,
			deploy.Environment,
			deploy.Ref,
			deploy.User.Name,
			(time.Duration(deploy.Duration) * time.Second).String(),
			renderer.FormatDate(deploy.FinishedAt),
			deploy.WebURL,
		})
	}
	table.Render()
}
