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

	"github.com/spf13/cobra"
)

func DeactivateCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate an extension",
		Long: `Deactivate a previously registered and activated extension on a Company or Project.

The cli context is used to select where the extension shall be deactivated.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}
			if o.EntityID == "" {
				return ErrRequiredExtensionID
			}

			scope := NewActivationScope(restConfig)

			extensibilityClient := New(client)
			err = extensibilityClient.Deactivate(cmd.Context(), restConfig.CompanyID, o.EntityID, scope)
			cobra.CheckErr(err)
			fmt.Printf("Successfully deactivated extension %s for %s\n", o.EntityID, scope)
			return nil
		},
	}

	addExtensionIDRequiredFlag(o, cmd)
	return cmd
}
