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
	"github.com/spf13/cobra"
)

const (
	getMarketplaceEndpoint        = "/api/backend/marketplace/%s"
	JSON                   string = "json"
	YAML                   string = "yaml"
)

const ()

var SupportedFormats = map[string]string{
	JSON: JSON,
	YAML: YAML,
}

// GetCmd return a new cobra command for getting a single marketplace resource
func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Get",
		Short: "Get marketplace item",
		Long:  `Get a single marketplace item by its ID`,
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

			return getMarketplaceResource(client, id, outputFormat)
		},
	}

	cmd.Flags().StringP("output", "o", "json", "Output format. Allowed values: json, yaml")

	return cmd
}

func getMarketplaceItemByID(client *client.APIClient, resourceID string) (*MarketplaceItem, error) {
	if len(resourceID) == 0 {
		return nil, fmt.Errorf("missing company id, please set one with the flag or context")
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

	var marketplaceItem *MarketplaceItem
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

	var format string

	if _, ok := SupportedFormats[outputFormat]; ok {
		format = outputFormat
	} else {
		return fmt.Errorf("invalid output format %s", outputFormat)
	}

	if format == JSON {
		json, err := marketplaceItem.MarshalMarketplaceItem()
		if err != nil {
			return err
		}
		fmt.Println(string(json))
	} else if format == YAML {
		yaml, err := marketplaceItem.MarshalMarketplaceItemYaml()
		if err != nil {
			return err
		}
		fmt.Println(string(yaml))
	}

	return nil
}
