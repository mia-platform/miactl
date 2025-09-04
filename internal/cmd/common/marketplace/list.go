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
	"strconv"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

type GetMarketplaceItemsOptions struct {
	CompanyID string
	Public    bool
	Page      int
}

func PrintMarketplaceItems(context context.Context, client *client.APIClient, options GetMarketplaceItemsOptions, p printer.IPrinter, endpoint string) error {
	marketplaceItems, err := fetchMarketplaceItems(context, client, options, endpoint)
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

func fetchMarketplaceItems(ctx context.Context, client *client.APIClient, options GetMarketplaceItemsOptions, endpoint string) ([]*resources.MarketplaceItem, error) {
	err := validateOptions(options)
	if err != nil {
		return nil, err
	}

	request := buildRequest(client, options, endpoint)
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
	requestedSpecificCompany := len(options.CompanyID) > 0

	if !requestedSpecificCompany {
		return marketplace.ErrMissingCompanyID
	}

	return nil
}

func buildRequest(client *client.APIClient, options GetMarketplaceItemsOptions, endpoint string) *client.Request {
	request := client.Get().APIPath(endpoint)
	switch {
	case options.Public:
		request.SetParam("includeTenantId", options.CompanyID)
	case !options.Public:
		request.SetParam("tenantId", options.CompanyID)
	}

	// marketplace command API does not support pagination
	if endpoint == "/api/marketplace/" {
		if options.Page <= 0 {
			request.SetParam("page", "1")
		} else {
			request.SetParam("page", strconv.Itoa(options.Page))
		}
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
