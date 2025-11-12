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
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	versionProjectCmdUsage = "version"
	versionProjectCmdShort = "Create a new version for a project"
	versionProjectCmdLong  = `Create a new version for the specified Project based on an existing revision.`
)

type versionProjectOptions struct {
	ProjectID          string
	TagName            string
	Ref                string
	Message            string
	ReleaseDescription string
}

// VersionCmd returns a cobra command for creating a project version
func VersionCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   versionProjectCmdUsage,
		Short: versionProjectCmdShort,
		Long:  versionProjectCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			cmdOptions := versionProjectOptions{
				ProjectID:          restConfig.ProjectID,
				TagName:            options.Tag,
				Ref:                options.Revision,
				Message:            options.Message,
				ReleaseDescription: options.ReleaseDescription,
			}

			return handleVersionProjectCmd(cmd.Context(), client, cmdOptions)
		},
	}

	flags := cmd.Flags()
	options.AddProjectFlags(flags)
	options.AddTagFlags(flags)
	options.AddRevisionFlags(flags)

	flags.StringVarP(&options.Message, "message", "m", "", "short description for the version")
	flags.StringVar(&options.ReleaseDescription, "release-description", "", "detailed release notes for the version")

	// Required flags
	if err := cmd.MarkFlagRequired("tag"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("revision"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("message"); err != nil {
		panic(err)
	}

	return cmd
}

func handleVersionProjectCmd(ctx context.Context, client *client.APIClient, options versionProjectOptions) error {
	err := validateVersionProjectOptions(options)
	if err != nil {
		return err
	}

	err = createVersion(ctx, client, options)
	if err != nil {
		return fmt.Errorf("failed to create project version: %w", err)
	}

	fmt.Printf("Project version '%s' created successfully\n", options.TagName)
	return nil
}

func validateVersionProjectOptions(options versionProjectOptions) error {
	if len(options.ProjectID) == 0 {
		return errors.New("missing project ID, please provide a project ID with the --project-id flag")
	}

	if len(options.TagName) == 0 {
		return errors.New("missing tag name, please provide a tag name with the --tag flag")
	}

	if len(options.Ref) == 0 {
		return errors.New("missing revision, please provide a revision with the --revision flag")
	}

	if len(options.Message) == 0 {
		return errors.New("missing message, please provide a message with the --message flag")
	}

	return nil
}

func createVersion(ctx context.Context, client *client.APIClient, options versionProjectOptions) error {
	// Create the request body
	requestBody := map[string]interface{}{
		"tagName":            options.TagName,
		"ref":                options.Ref,
		"message":            options.Message,
		"releaseDescription": options.ReleaseDescription,
	}

	body, err := resources.EncodeResourceToJSON(requestBody)
	if err != nil {
		return fmt.Errorf("cannot encode version request: %w", err)
	}

	// Create the endpoint
	endpoint := fmt.Sprintf("/api/backend/projects/%s/versions", options.ProjectID)
	
	response, err := client.
		Post().
		APIPath(endpoint).
		Body(body).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create project version: %w", err)
	}
	if err := response.Error(); err != nil {
		return err
	}

	return nil
}
