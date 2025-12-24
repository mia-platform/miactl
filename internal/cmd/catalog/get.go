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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/catalog"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
)

const (
	getItemByItemIDAndVersionEndpointTemplate = "/api/tenants/%s/marketplace/items/%s/versions/%s"

	cmdGetLongDescription = `Get a single Catalog item

	This command works with Mia-Platform Console v14.0.0 or later.

	You need to specify the itemId, via the respective flag. The company-id flag can be omitted if it is already set in the context.
	`
	cmdGetUse = "get { --item-id item-id --version version }"
)

// GetCmd return a new cobra command for getting a single catalog resource
func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdGetUse,
		Short: "Get Catalog item",
		Long:  cmdGetLongDescription,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), client, 14, 0)
			if versionError != nil {
				return versionError
			}
			if !canUseNewAPI {
				return catalog.ErrUnsupportedCompanyVersion
			}

			serializedItem, err := getItemEncodedWithFormat(
				cmd.Context(),
				client,
				restConfig.CompanyID,
				options.MarketplaceItemID,
				options.MarketplaceItemVersion,
				options.OutputFormat,
			)
			cobra.CheckErr(err)

			fmt.Println(serializedItem)
			return nil
		},
	}

	options.AddOutputFormatFlag(cmd.Flags(), encoding.JSON)

	itemIDFlagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	versionFlagName := options.AddMarketplaceVersionFlag(cmd.Flags())

	cmd.MarkFlagsRequiredTogether(itemIDFlagName, versionFlagName)

	return cmd
}

// getItemEncodedWithFormat retrieves the catalog item corresponding to the specified identifier, serialized with the specified outputFormat
func getItemEncodedWithFormat(ctx context.Context, client *client.APIClient, companyID, itemID, version, outputFormat string) (string, error) {
	if companyID == "" {
		return "", marketplace.ErrMissingCompanyID
	}
	endpoint := fmt.Sprintf(getItemByItemIDAndVersionEndpointTemplate, companyID, itemID, version)
	item, err := commonMarketplace.PerformGetItemRequest(ctx, client, endpoint)

	if err != nil {
		return "", err
	}

	data, err := item.Marshal(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
