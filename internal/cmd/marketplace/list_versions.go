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
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/spf13/cobra"
)

const listItemVersionsEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources/%s/versions"

var (
	ErrGenericServerError = errors.New("server error while fetching item versions")
)

// ListVersionCmd return a new cobra command for listing marketplace item versions
func ListVersionCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List versions of a Marketplace item (ALPHA)",
		Long: `List the currently available versions of a Marketplace item.
The command will output a table with each version of the item.

This command is in ALPHA state. This means that it can be subject to breaking changes in the next versions of miactl.`,
		Run: func(cmd *cobra.Command, _ []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			releases, err := getItemVersions(
				cmd.Context(),
				client,
				restConfig.CompanyID,
				options.MarketplaceItemID,
			)
			cobra.CheckErr(err)

			buildItemVersionListTable(releases, options.Printer(
				clioptions.DisableWrapLines(true),
			))
		},
	}

	flagName := options.AddMarketplaceItemIDFlag(cmd.Flags())
	err := cmd.MarkFlagRequired(flagName)
	if err != nil {
		// the error is only due to a programming error (missing command flag), hence panic
		panic(err)
	}

	return cmd
}

func getItemVersions(ctx context.Context, client *client.APIClient, companyID, itemID string) (*[]marketplace.Release, error) {
	if companyID == "" {
		return nil, marketplace.ErrMissingCompanyID
	}
	resp, err := client.
		Get().
		APIPath(
			fmt.Sprintf(listItemVersionsEndpointTemplate, companyID, itemID),
		).
		Do(ctx)

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
		return nil, fmt.Errorf("%w: %s", marketplace.ErrItemNotFound, itemID)
	}
	return nil, ErrGenericServerError
}

func buildItemVersionListTable(releases *[]marketplace.Release, p printer.IPrinter) {
	p.Keys("Version", "Name", "Description")

	for _, release := range *releases {
		description := "-"
		if release.Description != "" {
			description = release.Description
		}
		p.Record(
			release.Version,
			release.Name,
			description,
		)
	}
	p.Print()
}
