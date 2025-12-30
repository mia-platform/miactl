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
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
)

const (
	listVersionsCmdUsage = "list"
	listVersionsCmdShort = "List versions for a project"
	listVersionsCmdLong  = `List all versions for the specified Project.`
)

type listVersionsOptions struct {
	ProjectID    string
	OutputFormat string
}

// VersionListCmd returns a cobra command for listing project versions
func VersionListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   listVersionsCmdUsage,
		Short: listVersionsCmdShort,
		Long:  listVersionsCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			cmdOptions := listVersionsOptions{
				ProjectID:    restConfig.ProjectID,
				OutputFormat: options.OutputFormat,
			}

			return listProjectVersions(cmd.Context(), client, cmdOptions, cmd.OutOrStdout())
		},
	}

	flags := cmd.Flags()
	options.AddProjectFlags(flags)
	options.AddOutputFormatFlag(flags, "json")

	return cmd
}

func listProjectVersions(ctx context.Context, client *client.APIClient, options listVersionsOptions, writer io.Writer) error {
	if len(options.ProjectID) == 0 {
		return errors.New("missing project ID, please provide a project ID with the --project-id flag")
	}

	// Create the endpoint
	endpoint := fmt.Sprintf("/api/backend/projects/%s/versions", options.ProjectID)

	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to list project versions: %w", err)
	}
	if err := response.Error(); err != nil {
		return err
	}

	versions := make([]map[string]interface{}, 0)
	if err := response.ParseResponse(&versions); err != nil {
		return fmt.Errorf("cannot parse project versions: %w", err)
	}

	// If no versions are found, inform the user
	if len(versions) == 0 {
		fmt.Fprintln(writer, "No versions found for the project")
		return nil
	}

	// Format and output the versions
	bytes, err := encoding.MarshalData(versions, options.OutputFormat, encoding.MarshalOptions{Indent: true})
	if err != nil {
		return err
	}

	fmt.Fprintln(writer, string(bytes))
	return nil
}
