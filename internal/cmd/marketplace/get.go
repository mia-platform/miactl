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
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/spf13/cobra"
)

const (
	getItemByObjectIDEndpointTemplate         = "/api/backend/marketplace/%s"
	getItemByItemIDAndVersionEndpointTemplate = "/tenants/%s/resources/%s/versions/%s"
)

// GetCmd return a new cobra command for getting a single marketplace resource
func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get { --item-id item-id --version version } | --object-id object-id [FLAGS]...",
		Short: "Get Marketplace item",
		Long: `Get a single Marketplace item

You need to specify either:
- the itemId and the version, respectively  (recommended)
- the ObjectID of the item with the flag object-id
Passing the ObjectID is expected only when dealing with deprecated Marketplace items missing the itemId and/or version fields.
Otherwise, it is preferable to pass the tuple itemId-version.
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			var id string
			if len(args) > 0 {
				id = args[0]
			}

			serializedItem, err := getItemEncodedWithFormat(client, id, "", "", options.OutputFormat)
			cobra.CheckErr(err)

			fmt.Println(serializedItem)
			return nil
		},
	}

	options.AddOutputFormatFlag(cmd.Flags(), encoding.JSON)

	return cmd
}

func getItemByObjectID(client *client.APIClient, objectID string) (*marketplace.Item, error) {
	if len(objectID) == 0 {
		return nil, fmt.Errorf("missing resource id, please provide one")
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(getItemByObjectIDEndpointTemplate, objectID)).
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

// getItemEncodedWithFormat retrieves the marketplace item corresponding to the specified identifier, serialized with the specified outputFormat
func getItemEncodedWithFormat(client *client.APIClient, objectID, itemID, version, outputFormat string) (string, error) {
	marketplaceItem, err := getItemByObjectID(client, objectID)
	if err != nil {
		return "", err
	}

	data, err := marketplaceItem.MarshalItem(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
