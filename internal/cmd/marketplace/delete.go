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
	deleteMarketplaceEndpoint = "/api/backend/marketplace/tenants/%s/resources/%s"
)

// DeleteCmd return a new cobra command for deleting a single marketplace resource
func DeleteCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [resource-id]",
		Short: "Delete marketplace item",
		Long:  `Delete a single marketplace item by its ID`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			var resourceID string
			if len(args) > 0 {
				resourceID = args[0]
			}

			return deleteMarketplaceResource(client, restConfig.CompanyID, resourceID)
		},
	}

	return cmd
}

func deleteMarketplaceResource(client *client.APIClient, companyID string, resourceID string) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}

	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(deleteMarketplaceEndpoint, companyID, resourceID)).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	switch resp.StatusCode() {
	case 204:
		fmt.Println("resource deleted successfully")
		return nil
	case 404:
		return fmt.Errorf("resource not found")
	case 500:
		return fmt.Errorf("error while deleting resource")
	default:
		return fmt.Errorf("unexpected server response: %d", resp.StatusCode())
	}
}
