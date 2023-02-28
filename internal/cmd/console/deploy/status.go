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

package deploy

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/sdk"
	"github.com/mia-platform/miactl/old/sdk/deploy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewStatusCmd() *cobra.Command {
	var (
		baseURL         string
		apiToken        string
		projectID       string
		environment     string
		skipCertificate bool
		certificatePath string
	)

	cmd := &cobra.Command{
		Use:   "status deployId",
		Short: "verify status of deploy pipeline",
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			baseURL = viper.GetString("apibaseurl")
			apiToken = viper.GetString("apitoken")
			projectID = viper.GetString("project")

			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}
			if apiToken == "" {
				return errors.New("missing API token - please login")
			}
			if projectID == "" {
				return cmd.MarkFlagRequired("project")
			}

			// set these flag only in case they are defined
			skipCertificate, _ = cmd.Flags().GetBool("insecure")
			certificatePath = viper.GetString("ca-cert")

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := factory.FromContext(cmd.Context(), sdk.Options{
				APIBaseURL:            baseURL,
				APIToken:              apiToken,
				SkipCertificate:       skipCertificate,
				AdditionalCertificate: certificatePath,
			})
			if err != nil {
				return err
			}

			pipelineID, err := strconv.Atoi(args[0])
			if err != nil {
				f.Renderer.Error(fmt.Errorf("unexpected pipeline id: %w", err)).Render()
				return nil
			}

			result, err := f.MiaClient.Deploy.GetDeployStatus(projectID, pipelineID, environment)
			if err != nil {
				return err
			}

			visualizeStatusResponse(f, projectID, result)

			switch result.Status {
			case deploy.Failed:
				return fmt.Errorf("deploy pipeline failed")
			case deploy.Canceled:
				return fmt.Errorf("deploy pipeline canceled")
			default:
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "the environment where the project has been deployed")
	// Note: although this flag is defined as a persistent flag in the root command,
	// in order to be set during tests it must be defined also at command level
	cmd.Flags().BoolVar(&skipCertificate, "insecure", false, "whether to not check server certificate")

	return cmd
}

func visualizeStatusResponse(f *factory.Factory, projectID string, rs deploy.StatusResponse) {
	headers := []string{"Project Id", "Deploy Id", "Status"}
	table := f.Renderer.Table(headers)
	table.Append([]string{projectID, strconv.FormatInt(int64(rs.ID), 10), rs.Status})
	table.Render()
}
