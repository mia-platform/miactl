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
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/spf13/cobra"
)

const (
	describeProjectCmdUsage = "describe"
	describeProjectCmdShort = "Describe a Project configuration"
	describeProjectCmdLong  = `Describe the configuration of the specified Project.`
)

type describeProjectOptions struct {
	ProjectName  string
	RevisionName string
	VersionName  string
	OutputFormat string
}

// DescribeCmd returns a cobra command for describing a project configuration
func DescribeCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   describeProjectCmdUsage,
		Short: describeProjectCmdShort,
		Long:  describeProjectCmdLong,
		RunE: func(cmd *cobra.Command, _args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			cmdOptions := describeProjectOptions{
				RevisionName: options.Revision,
				VersionName:  options.Version,
				ProjectName:  restConfig.ProjectID,
				OutputFormat: options.OutputFormat,
			}

			return describeProject(cmd.Context(), client, cmdOptions)
		},
	}

	// add cmd flags
	flags := cmd.Flags()
	options.AddProjectFlags(flags)
	options.AddRevisionFlags(flags)
	options.AddVersionFlags(flags)
	options.AddOutputFormatFlag(flags, "json")

	return cmd
}

func describeProject(ctx context.Context, client *client.APIClient, options describeProjectOptions) error {
	if len(options.ProjectName) == 0 {
		return fmt.Errorf("missing project name, please provide a project name as argument")
	}

	ref, err := getConfigRef(options.RevisionName, options.VersionName)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/api/backend/projects/%s/%s/configuration", options.ProjectName, ref)
	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get project %s, ref %s: %w", options.ProjectName, ref, err)
	}
	if err := response.Error(); err != nil {
		return err
	}

	projectConfig := make(map[string]any, 0)
	if err := response.ParseResponse(&projectConfig); err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	bytes, err := encoding.MarshalData(projectConfig, options.OutputFormat, encoding.MarshalOptions{Indent: true})
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func getConfigRef(revisionName, versionName string) (string, error) {
	if len(revisionName) > 0 && len(versionName) > 0 {
		return "", fmt.Errorf("both revision and version specified, please provide only one")
	}

	if len(revisionName) > 0 {
		return fmt.Sprintf("revisions/%s", revisionName), nil
	}
	if len(versionName) > 0 {
		return fmt.Sprintf("versions/%s", versionName), nil
	}

	return "", fmt.Errorf("missing revision/version name, please provide one as argument")
}
