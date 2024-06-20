// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/spf13/cobra"
)

func addExtensionIDRequiredFlag(options *clioptions.CLIOptions, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&options.EntityID, "extension-id", "i", "", "the extension id that should be deleted")
}

// DeleteCmd return a new cobra command for listing companies
func DeletCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete extension",
		Long:  "Delete a previously register extension for the Company.",
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
			err = extensibilityClient.Delete(cmd.Context(), restConfig.CompanyID, options.EntityID)
			cobra.CheckErr(err)
			fmt.Println("Successfully deleted extension from Company")
			return nil
		},
	}

	addExtensionIDRequiredFlag(options, cmd)
	return cmd
}
