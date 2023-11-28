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

			list, err := buildMarketplaceItemsList(client, restConfig.CompanyID)
			cobra.CheckErr(err)

			fmt.Println(list)
		},
	}
}

func getMarketplaceItemsByCompanyID(client *client.APIClient, companyID string) ([]*resources.MarketplaceItem, error) {
	if len(companyID) == 0 {
		return nil, fmt.Errorf("missing company id, please set one with the flag or context")
	}

	resp, err := client.
		Get().
		SetParam("tenantId", companyID).
		APIPath(listMarketplaceEndpoint).
		Do(context.Background())

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

// buildMarketplaceItemsList retrieves the marketplace items belonging to the current context
// and returns a string with a human-readable list
func buildMarketplaceItemsList(client *client.APIClient, companyID string) (string, error) {
	marketplaceItems, err := getMarketplaceItemsByCompanyID(client, companyID)
	if err != nil {
		return "", err
	}
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAutoWrapText(true)
	//TODO: OBJECT ID
	table.SetHeader([]string{"ID", "Item ID", "Name", "Type"})
	for _, marketplaceItem := range marketplaceItems {
		table.Append([]string{
			marketplaceItem.ID,
			marketplaceItem.ItemID,
			marketplaceItem.Name,
			marketplaceItem.Type,
		})
	}
	table.Render()

	return strBuilder.String(), nil
}
