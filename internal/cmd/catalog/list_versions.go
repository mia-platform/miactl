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

package catalog

import (
	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/resources/catalog"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const listItemVersionsEndpointTemplate = "/api/tenants/%s/marketplace/items/%s/versions"

// ListVersionCmd return a new cobra command for listing marketplace item versions
func ListVersionCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List versions of a Marketplace item",
		Long: `List the currently available versions of a Marketplace item.
The command will output a table with each version of the item. It works with Mia-Platform Console v14.0.0 or later.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), client, 14, 0)
			if !canUseNewAPI || versionError != nil {
				return catalog.ErrUnsupportedCompanyVersion
			}

			releases, err := commonMarketplace.GetItemVersions(
				cmd.Context(),
				client,
				listItemVersionsEndpointTemplate,
				restConfig.CompanyID,
				options.MarketplaceItemID,
			)
			cobra.CheckErr(err)

			commonMarketplace.PrintItemVersionList(releases, options.Printer(
				clioptions.DisableWrapLines(true),
			))

			return nil
		},
	}

	flagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	err := cmd.MarkFlagRequired(flagName)
	if err != nil {
		// the error is only due to a programming error (missing command flag), hence panic
		panic(err)
	}

	return cmd
}
