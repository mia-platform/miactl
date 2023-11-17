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
	"errors"
	"fmt"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listItemVersionsEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions"
)

// ListVersionCmd return a new cobra command for listing marketplace item versions
func ListVersionCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List versions of a Marketplace item",
		Long: `List the currently available versions of a Marketplace item.
The command will output a table with the release notes of each specific version.`,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			releases, err := getItemVersions(
				client,
				options.CompanyID,
				options.MarketplaceItemID,
			)
			cobra.CheckErr(err)

			list, err := buildItemVersionList(releases)
			cobra.CheckErr(err)

			fmt.Println(list)
		},
	}

	options.AddMarketplaceGetItemVersionsFlags(cmd)

	return cmd
}

func getItemVersions(client *client.APIClient, companyID, itemID string) ([]marketplace.Release, error) {
	return nil, errors.New("not implemented")
}

// buildMarketplaceItemsList retrieves the marketplace items belonging to the current context
// and returns a string with a human-readable list
func buildItemVersionList(releases []marketplace.Release) (string, error) {
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAutoWrapText(true)
	table.SetHeader([]string{"Version", "Name", "Description"})
	for _, release := range releases {
		table.Append([]string{
			release.Version,
			release.Name,
			release.Description,
		})
	}
	table.Render()

	return strBuilder.String(), nil
}
