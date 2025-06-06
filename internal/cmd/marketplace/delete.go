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
	"errors"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	commonMarketplace "github.com/mia-platform/miactl/internal/cmd/common/marketplace"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	// deleteItemEndpointTemplate formatting template for item deletion by objectID backend endpoint; specify tenantID, objectID
	deleteItemEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s"
	// deleteItemByTupleEndpointTemplate formatting template for item deletion by the tuple itemID versionID endpoint; specify companyID, itemID, version
	deleteItemByTupleEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions/%s"

	cmdDeleteLongDescription = `Delete a single Marketplace item

	You need to specify either:
	- the companyId, itemId and version, via the respective flags (recommended). The company-id flag can be omitted if it is already set in the context.
	- the ObjectID of the item with the flag object-id

	Passing the ObjectID is expected only when dealing with deprecated Marketplace items missing the itemId and/or version fields.
	Otherwise, it is preferable to pass the tuple companyId-itemId-version.
	`
	cmdUse = "delete { --item-id item-id --version version } | --object-id object-id [flags]..."
)

// DeleteCmd return a new cobra command for deleting a single marketplace resource
func DeleteCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:        cmdUse,
		Short:      "Delete a Marketplace item",
		Long:       cmdDeleteLongDescription,
		SuggestFor: []string{"rm"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return marketplace.ErrMissingCompanyID
			}

			if options.MarketplaceItemObjectID != "" {
				err = deleteItemByObjectID(cmd.Context(), client, companyID, options.MarketplaceItemObjectID)
				cobra.CheckErr(err)
				return nil
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
		PostRun: util.CheckVersionAndShowMessage(options, 14, 0, marketplace.DeprecatedMessage),
	}

	itemObjectIDFlagName := options.AddMarketplaceItemObjectIDFlag(cmd.Flags())

	itemIDFlagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	versionFlagName := options.AddMarketplaceVersionFlag(cmd.Flags())

	cmd.MarkFlagsRequiredTogether(itemIDFlagName, versionFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, itemIDFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, versionFlagName)
	cmd.MarkFlagsOneRequired(itemObjectIDFlagName, itemIDFlagName, versionFlagName)

	return cmd
}

func deleteItemByObjectID(ctx context.Context, client *client.APIClient, companyID, objectID string) error {
	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteItemEndpointTemplate, companyID, objectID)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	return commonMarketplace.CheckDeleteResponseErrors(resp)
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
