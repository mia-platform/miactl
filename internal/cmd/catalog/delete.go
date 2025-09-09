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
	"errors"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/resources/catalog"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	// deleteItemByTupleEndpointTemplate formatting template for item deletion by the tuple itemID versionID endpoint; specify companyID, itemID, version
	deleteItemByTupleEndpointTemplate = "/api/tenants/%s/marketplace/items/%s/versions/%s"

	cmdDeleteLongDescription = `Delete a single Catalog item

	This command works with Mia-Platform Console v14.0.0 or later.

	You need to specify the companyId, itemId and version, via the respective flags (recommended). The company-id flag can be omitted if it is already set in the context.
	`
	cmdUse = "delete { --item-id item-id --version version }"
)

// DeleteCmd return a new cobra command for deleting a single catalog resource
func DeleteCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:        cmdUse,
		Short:      "Delete a Catalog item",
		Long:       cmdDeleteLongDescription,
		SuggestFor: []string{"rm"},
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

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return marketplace.ErrMissingCompanyID
			}

			if options.MarketplaceItemVersion != "" && options.MarketplaceItemID != "" {
				err = deleteItemByItemIDAndVersion(
					cmd.Context(),
					client,
					companyID,
					options.MarketplaceItemID,
					options.MarketplaceItemVersion,
				)
				cobra.CheckErr(err)
				return nil
			}

			return errors.New("invalid input parameters")
		},
	}

	itemIDFlagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	versionFlagName := options.AddMarketplaceVersionFlag(cmd.Flags())

	cmd.MarkFlagsRequiredTogether(itemIDFlagName, versionFlagName)

	return cmd
}

func deleteItemByItemIDAndVersion(ctx context.Context, client *client.APIClient, companyID, itemID, version string) error {
	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteItemByTupleEndpointTemplate, companyID, itemID, version)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	return commonMarketplace.CheckDeleteResponseErrors(resp)
}
