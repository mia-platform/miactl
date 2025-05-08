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
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	getItemByObjectIDEndpointTemplate         = "/api/backend/marketplace/%s"
	getItemByItemIDAndVersionEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions/%s"

	cmdGetLongDescription = `Get a single Marketplace item

	You need to specify either:
	- the companyId, itemId and version, via the respective flags (recommended). The company-id flag can be omitted if it is already set in the context.
	- the ObjectID of the item with the flag object-id

	Passing the ObjectID is expected only when dealing with deprecated Marketplace items missing the itemId and/or version fields.
	Otherwise, it is preferable to pass the tuple companyId-itemId-version.
	`
	cmdGetUse = "get { --item-id item-id --version version } | --object-id object-id [FLAGS]..."
)

// GetCmd return a new cobra command for getting a single marketplace resource
func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdGetUse,
		Short: "Get Marketplace item",
		Long:  cmdGetLongDescription,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			serializedItem, err := getItemEncodedWithFormat(
				cmd.Context(),
				client,
				options.MarketplaceItemObjectID,
				restConfig.CompanyID,
				options.MarketplaceItemID,
				options.MarketplaceItemVersion,
				options.OutputFormat,
			)
			cobra.CheckErr(err)

			fmt.Println(serializedItem)
			return nil
		},
		PostRun: util.CheckVersionAndShowMessage(options, 14, 0, marketplace.DeprecatedMessage),
	}

	options.AddOutputFormatFlag(cmd.Flags(), encoding.JSON)

	itemObjectIDFlagName := options.AddMarketplaceItemObjectIDFlag(cmd.Flags())

	itemIDFlagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	versionFlagName := options.AddMarketplaceVersionFlag(cmd.Flags())

	cmd.MarkFlagsRequiredTogether(itemIDFlagName, versionFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, itemIDFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, versionFlagName)
	cmd.MarkFlagsOneRequired(itemObjectIDFlagName, itemIDFlagName, versionFlagName)

	return cmd
}
func getItemByObjectID(ctx context.Context, client *client.APIClient, objectID string) (*marketplace.Item, error) {
	return commonMarketplace.PerformGetItemRequest(ctx, client, fmt.Sprintf(getItemByObjectIDEndpointTemplate, objectID))
}

func getItemByItemIDAndVersion(ctx context.Context, client *client.APIClient, companyID, itemID, version string) (*marketplace.Item, error) {
	endpoint := fmt.Sprintf(getItemByItemIDAndVersionEndpointTemplate, companyID, itemID, version)
	return commonMarketplace.PerformGetItemRequest(ctx, client, endpoint)
}

// getItemEncodedWithFormat retrieves the marketplace item corresponding to the specified identifier, serialized with the specified outputFormat
func getItemEncodedWithFormat(ctx context.Context, client *client.APIClient, objectID, companyID, itemID, version, outputFormat string) (string, error) {
	var item *marketplace.Item
	var err error
	if objectID != "" {
		item, err = getItemByObjectID(ctx, client, objectID)
	} else {
		if companyID == "" {
			return "", marketplace.ErrMissingCompanyID
		}
		item, err = getItemByItemIDAndVersion(ctx, client, companyID, itemID, version)
	}
	if err != nil {
		return "", err
	}

	data, err := item.Marshal(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
