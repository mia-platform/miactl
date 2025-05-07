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

package marketplace

import (
	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	listMarketplaceEndpoint = "/api/backend/marketplace/"
	listCmdLong             = `List Marketplace items

    This command lists the Marketplace items of a company.

    you can also specify the following flags:
    - --public - if this flag is set, the command fetches not only the items from the requested company, but also the public Marketplace items from other companies.
    `
	listCmdUse = "list --company-id company-id"
)

// ListCmd return a new cobra command for listing marketplace items
func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     listCmdUse,
		Short:   "List marketplace items",
		Long:    listCmdLong,
		Run:     runListCmd(options),
		PostRun: util.CheckVersionAndShowMessage(options, 14, 0, marketplace.DeprecatedMessage),
	}

	options.AddPublicFlag(cmd.Flags())

	return cmd
}

func runListCmd(options *clioptions.CLIOptions) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		restConfig, err := options.ToRESTConfig()
		cobra.CheckErr(err)
		apiClient, err := client.APIClientForConfig(restConfig)
		cobra.CheckErr(err)

		marketplaceItemsOptions := commonMarketplace.GetMarketplaceItemsOptions{
			CompanyID: restConfig.CompanyID,
			Public:    options.MarketplaceFetchPublicItems,
		}

		err = commonMarketplace.PrintMarketplaceItems(cmd.Context(), apiClient, marketplaceItemsOptions, options.Printer(), listMarketplaceEndpoint)
		cobra.CheckErr(err)
	}
}
