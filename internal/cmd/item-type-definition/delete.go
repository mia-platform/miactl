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

package itd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

var (
	ErrServerDeleteItem     = errors.New("server error while deleting item type definition")
	ErrUnexpectedDeleteItem = errors.New("unexpected response while deleting item")
)

const (
	deleteItdEndpoint = "/api/tenants/%s/marketplace/item-type-definitions/%s/"

	cmdDeleteLongDescription = `Delete an Item Type Definition. It works with Mia-Platform Console v14.1.0 or later.

	You need to specify the companyId and the item type definition name via the respective flags (recommended). The company-id flag can be omitted if it is already set in the context.
	`
	deleteCmdUse = "delete --name name --version version"
)

func DeleteCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:        deleteCmdUse,
		Short:      "Delete an Item Type Definition",
		Long:       cmdDeleteLongDescription,
		SuggestFor: []string{"rm"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), client, 14, 1)
			if versionError != nil {
				return versionError
			}
			if !canUseNewAPI {
				return itd.ErrUnsupportedCompanyVersion
			}

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return itd.ErrMissingCompanyID
			}

			if options.MarketplaceItemVersion != "" && options.MarketplaceItemID != "" {
				err = deleteITD(
					cmd.Context(),
					client,
					companyID,
					options.ItemTypeDefinitionName,
				)
				cobra.CheckErr(err)
				return nil
			}

			return errors.New("invalid input parameters")
		},
	}

	ITDFlagName := options.AddItemTypeDefinitionNameFlag(cmd.Flags())

	err := cmd.MarkFlagRequired(ITDFlagName)
	if err != nil {
		// the error is only due to a programming error (missing command flag), hence panic
		panic(err)
	}

	return cmd
}

func deleteITD(ctx context.Context, client *client.APIClient, companyID, name string) error {
	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteItdEndpoint, companyID, name)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		fmt.Println("item deleted successfully")
		return nil
	case http.StatusNotFound:
		return itd.ErrItemNotFound
	default:
		if resp.StatusCode() >= http.StatusInternalServerError {
			return ErrServerDeleteItem
		}
		return fmt.Errorf("%w: %d", ErrUnexpectedDeleteItem, resp.StatusCode())
	}
}
