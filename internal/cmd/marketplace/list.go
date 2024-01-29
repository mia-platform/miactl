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
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listMarketplaceEndpoint = "/api/backend/marketplace/"
)

// ListCmd return a new cobra command for listing marketplace items
func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List marketplace items",
		Long:  `List the Marketplace items that the current user can access.`,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			getMarketplaceItemsOptions := GetMarketplaceItemsOptions{
				companyID: options.CompanyID,
				public:    false,
			}

			table, err := getMarketplaceItemsTable(context.Background(), client, getMarketplaceItemsOptions)
			cobra.CheckErr(err)

			fmt.Println(table)
		},
	}
}

type GetMarketplaceItemsOptions struct {
	companyID string
	public    bool
}

func getMarketplaceItemsTable(context context.Context, client *client.APIClient, options GetMarketplaceItemsOptions) (string, error) {

	marketplaceItems, err := getMarketplaceItemsByCompanyID(context, client, options)
	if err != nil {
		return "", err
	}

	table := buildMarketplaceItemsTable(marketplaceItems)
	return table, nil
}

func getMarketplaceItemsByCompanyID(ctx context.Context, client *client.APIClient, options GetMarketplaceItemsOptions) ([]*resources.MarketplaceItem, error) {
	if len(options.companyID) == 0 {
		return nil, marketplace.ErrMissingCompanyID
	}

	resp, err := client.
		Get().
		SetParam("tenantId", options.companyID).
		APIPath(listMarketplaceEndpoint).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	marketplaceItems := make([]*resources.MarketplaceItem, 0)
	if err := resp.ParseResponse(&marketplaceItems); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	return marketplaceItems, nil
}

func buildMarketplaceItemsTable(marketplaceItems []*resources.MarketplaceItem) string {
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAutoWrapText(true)
	table.SetHeader([]string{"Object ID", "Item ID", "Name", "Type", "Company ID"})
	for _, marketplaceItem := range marketplaceItems {
		table.Append([]string{
			marketplaceItem.ID,
			marketplaceItem.ItemID,
			marketplaceItem.Name,
			marketplaceItem.Type,
			marketplaceItem.TenantID,
		})
	}
	table.Render()

	return strBuilder.String()
}
