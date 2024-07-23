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

package extensions

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"

	"github.com/spf13/cobra"
)

// GetOneCmd return a new cobra command to get a single extension
func GetOneCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a registered extension",
		Long:  "Get details for a single registered extension for the company.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}
			if options.EntityID == "" {
				return ErrRequiredExtensionID
			}

			extensibilityClient := New(client)
			extension, err := extensibilityClient.GetOne(cmd.Context(), restConfig.CompanyID, options.EntityID)
			cobra.CheckErr(err)

			serialized, err := encoding.MarshalData(extension, options.OutputFormat, encoding.MarshalOptions{Indent: true})
			if err != nil {
				return err
			}
			fmt.Println(string(serialized))
			return nil
		},
	}

	options.AddOutputFormatFlag(cmd.Flags(), encoding.JSON)
	addExtensionIDFlag(options, cmd, "the extension id that should be retrieved")
	return cmd
}
