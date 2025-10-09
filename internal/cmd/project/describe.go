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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/configuration"
	"github.com/spf13/cobra"
)

const (
	describeProjectCmdUsage = "describe"
	describeProjectCmdShort = "Describe a Project configuration"
	describeProjectCmdLong  = `Describe the configuration of the specified Project.`

	ErrMultipleIdentifiers = "multiple identifiers specified, please provide only one"
	ErrMissingIdentifier   = "missing revision/version/branch/tag name, please provide one as argument"
)

type describeProjectOptions struct {
	RevisionName string
	VersionName  string
	BranchName   string
	TagName      string
	ProjectID    string
	OutputFormat string
}

// DescribeCmd returns a cobra command for describing a project configuration
func DescribeCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   describeProjectCmdUsage,
		Short: describeProjectCmdShort,
		Long:  describeProjectCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			cmdOptions := describeProjectOptions{
				RevisionName: options.Revision,
				VersionName:  options.Version,
				BranchName:   options.Branch,
				TagName:      options.Tag,
				ProjectID:    restConfig.ProjectID,
				OutputFormat: options.OutputFormat,
			}

			return describeProject(cmd.Context(), client, cmdOptions, cmd.OutOrStdout())
		},
	}

	flags := cmd.Flags()
	options.AddProjectFlags(flags)
	options.AddRevisionFlags(flags)
	options.AddVersionFlags(flags)
	options.AddBranchFlags(flags)
	options.AddTagFlags(flags)
	options.AddOutputFormatFlag(flags, "json")

	return cmd
}

func describeProject(ctx context.Context, client *client.APIClient, options describeProjectOptions, writer io.Writer) error {
	if len(options.ProjectID) == 0 {
		return errors.New("missing project name, please provide a project name as argument")
	}

	ref, err := GetRefFromOptions(options)
	if err != nil {
		return err
	}

	endpoint := ref.ConfigurationEndpoint(options.ProjectID)
	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get project %s, ref %s: %w", options.ProjectID, ref, err)
	}
	if err := response.Error(); err != nil {
		return err
	}

	projectConfig := make(map[string]any, 0)
	if err := response.ParseResponse(&projectConfig); err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	structuredConfig, err := configuration.BuildDescribeFromFlatConfiguration(projectConfig)
	if err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	bytes, err := encoding.MarshalData(structuredConfig, options.OutputFormat, encoding.MarshalOptions{Indent: true})
	if err != nil {
		return err
	}

	fmt.Fprintln(writer, string(bytes))
	return nil
}

func GetRefFromOptions(options describeProjectOptions) (configuration.Ref, error) {
	refType := ""
	refName := ""

	if len(options.RevisionName) > 0 {
		refType = configuration.RevisionRefType
		refName = options.RevisionName
	}

	if len(options.VersionName) > 0 {
		if len(refType) > 0 {
			return configuration.Ref{}, errors.New(ErrMultipleIdentifiers)
		}

		refType = configuration.VersionRefType
		refName = options.VersionName
	}

	if len(options.BranchName) > 0 {
		if len(refType) > 0 {
			return configuration.Ref{}, errors.New(ErrMultipleIdentifiers)
		}

		refType = configuration.BranchRefType
		refName = options.BranchName
	}

	if len(options.TagName) > 0 {
		if len(refType) > 0 {
			return configuration.Ref{}, errors.New(ErrMultipleIdentifiers)
		}

		refType = configuration.TagRefType
		refName = options.TagName
	}

	if len(refType) == 0 {
		return configuration.Ref{}, errors.New(ErrMissingIdentifier)
	}

	return configuration.NewRef(refType, refName)
}
