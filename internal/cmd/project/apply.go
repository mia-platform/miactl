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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/files"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/configuration"

	"github.com/spf13/cobra"
)

const (
	applyProjectCmdUsage = "apply"
	applyProjectCmdShort = "Apply a Project configuration"
	applyProjectCmdLong  = `Apply a Project configuration from a file.

The configuration file should contain a complete project configuration in JSON or YAML format.
This command will replace the current project configuration with the one provided in the file.
`
)

type applyProjectOptions struct {
	ProjectID    string
	RevisionName string
	FilePath     string
	Title        string
}

// ApplyCmd returns a cobra command for applying a project configuration
func ApplyCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   applyProjectCmdUsage,
		Short: applyProjectCmdShort,
		Long:  applyProjectCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			cmdOptions := applyProjectOptions{
				RevisionName: options.Revision,
				ProjectID:    restConfig.ProjectID,
				FilePath:     options.InputFilePath,
				Title:        options.Message,
			}

			return handleApplyProjectConfigurationCmd(cmd.Context(), client, cmdOptions)
		},
	}

	flags := cmd.Flags()
	options.AddProjectFlags(flags)
	options.AddRevisionFlags(flags)

	flags.StringVarP(&options.Message, "message", "m", "", "the message to use when saving the configuration")

	// file path flag is required
	flags.StringVarP(&options.InputFilePath, "file", "f", "", "path to JSON/YAML file containing the project configuration")
	if err := cmd.MarkFlagRequired("file"); err != nil {
		panic(err)
	}

	return cmd
}

func handleApplyProjectConfigurationCmd(ctx context.Context, client *client.APIClient, options applyProjectOptions) error {
	err := validateApplyProjectOptions(options)
	if err != nil {
		return err
	}

	err = applyConfiguration(ctx, client, options)
	if err != nil {
		return fmt.Errorf("failed to apply project configuration: %w", err)
	}

	fmt.Println("Project configuration applied successfully")
	return nil
}

func validateApplyProjectOptions(options applyProjectOptions) error {
	if len(options.ProjectID) == 0 {
		return errors.New("missing project name, please provide a project name as argument")
	}

	if len(options.FilePath) == 0 {
		return errors.New("missing file path, please provide a file path with the -f flag")
	}

	if len(options.RevisionName) == 0 {
		return errors.New("missing revision name, please provide a revision name")
	}

	return nil
}

func applyConfiguration(ctx context.Context, client *client.APIClient, options applyProjectOptions) error {
	ref, err := configuration.NewRef(configuration.RevisionRefType, options.RevisionName)
	if err != nil {
		return err
	}

	projectConfig := make(map[string]any)
	if err := files.ReadFile(options.FilePath, &projectConfig); err != nil {
		return fmt.Errorf("failed to read project configuration file: %w", err)
	}

	structuredConfig, err := configuration.BuildDescribeConfiguration(projectConfig)
	if err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	applyConfig := configuration.BuildApplyRequest(structuredConfig)

	if options.Title != "" {
		applyConfig = applyConfig.WithTitle(options.Title)
	}

	body, err := resources.EncodeResourceToJSON(applyConfig)
	if err != nil {
		return fmt.Errorf("cannot encode project configuration: %w", err)
	}

	endpoint := ref.ConfigurationEndpoint(options.ProjectID)
	response, err := client.
		Post().
		APIPath(endpoint).
		Body(body).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to apply project configuration: %w", err)
	}
	if err := response.Error(); err != nil {
		return err
	}

	return nil
}
