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
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/catalog"
	"github.com/spf13/cobra"
)

const (
	listMarketplaceEndpoint = "/api/marketplace/"
	listCmdLong             = `List Catalog items

    This command lists the Catalog items of a company.

    you can also specify the following flags:
    - --public - if this flag is set, the command fetches not only the items from the requested company, but also the public Catalog items from other companies.
    `
	listCmdUse = "list --company-id company-id"
)

// ListCmd return a new cobra command for listing catalog items
func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   listCmdUse,
		Short: "List catalog items",
		Long:  listCmdLong,
		Run:   runListCmd(options),
	}

	options.AddPublicFlag(cmd.Flags())

	return cmd
}

func runListCmd(options *clioptions.CLIOptions) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		restConfig, err := options.ToRESTConfig()
		cobra.CheckErr(err)
		apiClient, err := client.APIClientForConfig(restConfig)
		cobra.CheckErr(err)

		marketplaceItemsOptions := GetMarketplaceItemsOptions{
			companyID: restConfig.CompanyID,
			public:    options.MarketplaceFetchPublicItems,
		}

		err = printMarketplaceItems(cmd.Context(), apiClient, marketplaceItemsOptions, options.Printer())
		cobra.CheckErr(err)
	}
}

type GetMarketplaceItemsOptions struct {
	companyID string
	public    bool
}

func printMarketplaceItems(context context.Context, client *client.APIClient, options GetMarketplaceItemsOptions, p printer.IPrinter) error {
	marketplaceItems, err := fetchMarketplaceItems(context, client, options)
	if err != nil {
		return err
	}

	p.Keys("Object ID", "Item ID", "Name", "Type", "Company ID")
	for _, marketplaceItem := range marketplaceItems {
		p.Record(
			marketplaceItem.ID,
			marketplaceItem.ItemID,
			marketplaceItem.Name,
			marketplaceItem.Type,
			marketplaceItem.TenantID,
		)
	}
	p.Print()
	return nil
}

func fetchMarketplaceItems(ctx context.Context, client *client.APIClient, options GetMarketplaceItemsOptions) ([]*resources.MarketplaceItem, error) {
	err := validateOptions(options)
	if err != nil {
		return nil, err
	}

	request := buildRequest(client, options)
	resp, err := executeRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	marketplaceItems, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return marketplaceItems, nil
}

func validateOptions(options GetMarketplaceItemsOptions) error {
	requestedSpecificCompany := len(options.companyID) > 0

	if !requestedSpecificCompany {
		return catalog.ErrMissingCompanyID
	}

	return nil
}

func buildRequest(client *client.APIClient, options GetMarketplaceItemsOptions) *client.Request {
	request := client.Get().APIPath(listMarketplaceEndpoint)
	switch {
	case options.public:
		request.SetParam("includeTenantId", options.companyID)
	case !options.public:
		request.SetParam("tenantId", options.companyID)
	}

	return request
}

func executeRequest(ctx context.Context, request *client.Request) (*client.Response, error) {
	resp, err := request.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	if err := resp.Error(); err != nil {
		return nil, err
	}

	return resp, nil
}

func parseResponse(resp *client.Response) ([]*resources.MarketplaceItem, error) {
	marketplaceItems := make([]*resources.MarketplaceItem, 0)
	if err := resp.ParseResponse(&marketplaceItems); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	return marketplaceItems, nil
}
