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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/catalog"
	"github.com/spf13/cobra"
)

const (
	getItemByItemIDAndVersionEndpointTemplate = "/api/tenants/%s/marketplace/items/%s/versions/%s"

	cmdGetLongDescription = `Get a single Catalog item

	You need to specify the companyId, itemId and version, via the respective flags. The company-id flag can be omitted if it is already set in the context.
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

func getItemByItemIDAndVersion(ctx context.Context, client *client.APIClient, companyID, itemID, version string) (*catalog.Item, error) {
	endpoint := fmt.Sprintf(getItemByItemIDAndVersionEndpointTemplate, companyID, itemID, version)
	return performGetItemRequest(ctx, client, endpoint)
}

func performGetItemRequest(ctx context.Context, client *client.APIClient, endpoint string) (*catalog.Item, error) {
	resp, err := client.Get().APIPath(endpoint).Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	var marketplaceItem *catalog.Item
	if err := resp.ParseResponse(&marketplaceItem); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	if marketplaceItem == nil {
		return nil, fmt.Errorf("no catalog item returned in the response")
	}

	return marketplaceItem, nil
}

// getItemEncodedWithFormat retrieves the catalog item corresponding to the specified identifier, serialized with the specified outputFormat
func getItemEncodedWithFormat(ctx context.Context, client *client.APIClient, companyID, itemID, version, outputFormat string) (string, error) {
	var item *catalog.Item
	var err error
	if companyID == "" {
		return "", catalog.ErrMissingCompanyID
	}
	item, err = getItemByItemIDAndVersion(ctx, client, companyID, itemID, version)

	if err != nil {
		return "", err
	}

	data, err := item.Marshal(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
