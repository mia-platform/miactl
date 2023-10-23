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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/spf13/cobra"
)

const (
	getMarketplaceEndpoint = "/api/backend/marketplace/%s"
)

// GetCmd return a new cobra command for getting a single marketplace resource
func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get resource-id",
		Short: "Get Marketplace item",
		Long:  `Get a single Marketplace item by its ID`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			var id string
			if len(args) > 0 {
				id = args[0]
			}

			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			err = getMarketplaceResource(client, id, outputFormat)
			cobra.CheckErr(err)

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "json", "Output format. Allowed values: json, yaml")

	return cmd
}

func getMarketplaceItemByID(client *client.APIClient, resourceID string) (*marketplace.Item, error) {
	if len(resourceID) == 0 {
		return nil, fmt.Errorf("missing resource id, please provide one")
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(getMarketplaceEndpoint, resourceID)).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	var marketplaceItem *marketplace.Item
	if err := resp.ParseResponse(&marketplaceItem); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	if marketplaceItem == nil {
		return nil, fmt.Errorf("no marketplace item returned in the response")
	}

	return marketplaceItem, nil
}

// getMarketplaceResource retrieves the marketplace items for a given resource ID
func getMarketplaceResource(client *client.APIClient, resourceID string, outputFormat string) error {
	marketplaceItem, err := getMarketplaceItemByID(client, resourceID)
	if err != nil {
		return err
	}

	data, err := marketplaceItem.MarshalItem(outputFormat)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
