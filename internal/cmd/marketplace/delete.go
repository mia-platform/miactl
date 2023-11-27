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
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/spf13/cobra"
)

const (
	// deleteMarketplaceEndpointTemplate formatting template for item deletion by objectID backend endpoint; specify tenantID, objectID
	deleteMarketplaceEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s"
	// deleteItemByTupleEndpointTemplate formatting template for item deletion by the tuple itemID versionID endpoint; specify companyID, itemID, version
	deleteItemByTupleEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions/%s"
)

var (
	errServerDeleteItem     = errors.New("server error while deleting item")
	errUnexpectedDeleteItem = errors.New("unexpected response while deleting item")
)

// DeleteCmd return a new cobra command for deleting a single marketplace resource
func DeleteCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ",
		Short: "Delete Marketplace item",
		Long: `Delete a single Marketplace item by its ID
You need to specify either:
- the itemId and the version with the respective flags
- the ObjectID of the item with the flag object-id
`,
		SuggestFor: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return fmt.Errorf("missing company id, please set one with the flag or context")
			}

			if options.MarketplaceItemObjectID != "" {
				err = deleteItemByObjectID(cmd.Context(), client, companyID, options.MarketplaceItemObjectID)
				cobra.CheckErr(err)
				return nil
			}

			if options.MarketplaceItemVersion != "" && options.MarketplaceItemItemID != "" {
				err = deleteItemByItemIDAndVersion(
					cmd.Context(),
					client,
					companyID,
					options.MarketplaceItemItemID,
					options.MarketplaceItemVersion,
				)
				cobra.CheckErr(err)
				return nil
			}

			return errors.New("invalid input parameters")
		},
	}

	itemIDFlagName := options.AddMarketplaceItemItemIDFlag(cmd.Flags())
	versionFlagName := options.AddMarketplaceVersionFlag(cmd.Flags())
	itemObjectIDFlagName := options.AddMarketplaceItemObjectIDFlag(cmd.Flags())

	cmd.MarkFlagsRequiredTogether(itemIDFlagName, versionFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, itemIDFlagName)
	cmd.MarkFlagsMutuallyExclusive(itemObjectIDFlagName, versionFlagName)
	cmd.MarkFlagsOneRequired(itemObjectIDFlagName, itemIDFlagName, versionFlagName)

	return cmd
}

func deleteItemByObjectID(ctx context.Context, client *client.APIClient, companyID, objectID string) error {
	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteMarketplaceEndpointTemplate, companyID, objectID)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	return checkDeleteResponseErrors(resp)
}

func deleteItemByItemIDAndVersion(ctx context.Context, client *client.APIClient, companyID, itemID, version string) error {
	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteItemByTupleEndpointTemplate, companyID, itemID, version)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	return checkDeleteResponseErrors(resp)
}

func checkDeleteResponseErrors(resp *client.Response) error {
	switch resp.StatusCode() {
	case http.StatusNoContent:
		fmt.Println("item deleted successfully")
		return nil
	case http.StatusNotFound:
		return marketplace.ErrItemNotFound
	default:
		if resp.StatusCode() >= http.StatusInternalServerError {
			return errServerDeleteItem
		}
		return fmt.Errorf("%w: %d", errUnexpectedDeleteItem, resp.StatusCode())
	}
}
