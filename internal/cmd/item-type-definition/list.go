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
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/mia-platform/miactl/internal/util"
)

type GetItdsOptions struct {
	CompanyID string
	Public    bool
	Page      int
}

const (
	listItdEndpoint = "/api/marketplace/item-type-definitions/"
	listCmdLong     = `List Item Type Definitions

    This command lists the Item Type Definitions of a company. It works with Mia-Platform Console v14.1.0 or later.

		Results are paginated. By default, only the first page is shown.

    you can also specify the following flags:
    - --public - if this flag is set, the command fetches not only the items from the requested company, but also the public Catalog items from other companies.
		- --page - specify the page to fetch, default is 1
    `
	listCmdUse = "list --company-id company-id"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   listCmdUse,
		Short: "List item type definitions",
		Long:  listCmdLong,
		RunE:  runListCmd(options),
	}

	options.AddPublicFlag(cmd.Flags())
	options.AddPageFlag(cmd.Flags())

	return cmd
}

func runListCmd(options *clioptions.CLIOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		restConfig, err := options.ToRESTConfig()
		cobra.CheckErr(err)
		apiClient, err := client.APIClientForConfig(restConfig)
		cobra.CheckErr(err)

		canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), apiClient, 14, 1)
		if versionError != nil {
			return versionError
		}
		if !canUseNewAPI {
			return itd.ErrUnsupportedCompanyVersion
		}

		listItemsOptions := GetItdsOptions{
			CompanyID: restConfig.CompanyID,
			Public:    options.MarketplaceFetchPublicItems,
			Page:      options.Page,
		}

		err = PrintItds(cmd.Context(), apiClient, listItemsOptions, options.Printer(), listItdEndpoint)
		cobra.CheckErr(err)

		return nil
	}
}

func PrintItds(context context.Context, client *client.APIClient, options GetItdsOptions, p printer.IPrinter, endpoint string) error {
	itds, err := fetchItds(context, client, options, endpoint)
	if err != nil {
		return err
	}

	p.Keys("Name", "Display Name", "Visibility", "Publisher", "Versioning Supported")
	for _, itd := range itds {
		publisher := itd.Metadata.Publisher.Name
		if publisher == "" {
			publisher = "-"
		}

		p.Record(
			itd.Metadata.Name,
			itd.Metadata.DisplayName,
			itd.Metadata.Visibility.Scope,
			publisher,
			strconv.FormatBool(itd.Spec.IsVersioningSupported),
		)
	}
	p.Print()
	return nil
}

func fetchItds(ctx context.Context, client *client.APIClient, options GetItdsOptions, endpoint string) ([]*itd.ItemTypeDefinition, error) {
	err := validateOptions(options)
	if err != nil {
		return nil, err
	}

	request := buildRequest(client, options, endpoint)
	resp, err := executeRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	listItems, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	return listItems, nil
}

func validateOptions(options GetItdsOptions) error {
	requestedSpecificCompany := len(options.CompanyID) > 0

	if !requestedSpecificCompany {
		return itd.ErrMissingCompanyID
	}

	return nil
}

func buildRequest(client *client.APIClient, options GetItdsOptions, endpoint string) *client.Request {
	request := client.Get().APIPath(endpoint)
	switch {
	case options.Public:
		request.SetParam("visibility", "console"+","+options.CompanyID)
	case !options.Public:
		request.SetParam("visibility", options.CompanyID)
	}

	if options.Page <= 0 {
		request.SetParam("page", "1")
	} else {
		request.SetParam("page", strconv.Itoa(options.Page))
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

func parseResponse(resp *client.Response) ([]*itd.ItemTypeDefinition, error) {
	listItems := make([]*itd.ItemTypeDefinition, 0)
	if err := resp.ParseResponse(&listItems); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	return listItems, nil
}
