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
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const listItemVersionsEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions"

var (
	ErrItemVersionNotFound = errors.New("an item with the specified itemID wasn't found")
	ErrGenericItemVersion  = errors.New("server error while fetching item version")
	ErrMissingCompanyID    = errors.New("companyID is required")
)

// ListVersionCmd return a new cobra command for listing marketplace item versions
func ListVersionCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List versions of a Marketplace item",
		Long: `List the currently available versions of a Marketplace item.
The command will output a table with each version of the item.`,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			releases, err := getItemVersions(
				client,
				restConfig.CompanyID,
				options.MarketplaceItemID,
			)
			cobra.CheckErr(err)

			table := buildItemVersionList(releases)

			fmt.Println(table)
		},
	}

	options.AddMarketplaceGetItemVersionsFlags(cmd)

	return cmd
}

func getItemVersions(client *client.APIClient, companyID, itemID string) (*[]marketplace.Release, error) {
	if companyID == "" {
		return nil, ErrMissingCompanyID
	}
	resp, err := client.
		Get().
		APIPath(
			fmt.Sprintf(listItemVersionsEndpointTemplate, companyID, itemID),
		).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		releases := &[]marketplace.Release{}
		err = resp.ParseResponse(releases)
		if err != nil {
			return nil, fmt.Errorf("error parsing response body: %w", err)
		}
		return releases, nil
	case http.StatusNotFound:
		return nil, ErrItemVersionNotFound
	}
	return nil, ErrGenericItemVersion
}

// buildMarketplaceItemsList retrieves the marketplace items belonging to the current context
// and returns a string with a human-readable list
func buildItemVersionList(releases *[]marketplace.Release) string {
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Version", "Name", "Description"})

	for _, release := range *releases {
		description := "-"
		if release.Description != "" {
			description = release.Description
		}
		table.Append([]string{
			release.Version,
			release.Name,
			description,
		})
	}
	table.Render()

	return strBuilder.String()
}
