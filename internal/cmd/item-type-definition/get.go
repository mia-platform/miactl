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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	getItdEndpoint = "/api/tenants/%s/marketplace/item-type-definitions/%s/"
	getCmdLong     = `Get an Item Type Definition

   This command get an Item Type Definitions based on its name and tenant namespace. It works with Mia-Platform Console v14.1.0 or later.

   You need to specify the name via the respective flag. The company-id flag can be omitted if it is already set in the context and it is used as tenantId of the item type definition.
   `
	getCmdUse = "get --name name"
)

func GetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   getCmdUse,
		Short: "Get item type definition",
		Long:  getCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), client, 14, 1)
			if versionError != nil {
				return versionError
			}
			if !canUseNewAPI {
				return itd.ErrUnsupportedCompanyVersion
			}

			serializedItem, err := getItemEncodedWithFormat(
				cmd.Context(),
				client,
				restConfig.CompanyID,
				options.ItemTypeDefinitionName,
				options.OutputFormat,
			)
			cobra.CheckErr(err)

			fmt.Println(serializedItem)
			return nil
		},
	}

	nameFlagName := options.AddItemTypeDefinitionNameFlag(cmd.Flags())

	err := cmd.MarkFlagRequired(nameFlagName)
	if err != nil {
		// the error is only due to a programming error (missing command flag), hence panic
		panic(err)
	}

	return cmd
}

func getItemEncodedWithFormat(ctx context.Context, client *client.APIClient, companyID, name, outputFormat string) (string, error) {
	if companyID == "" {
		return "", itd.ErrMissingCompanyID
	}
	endpoint := fmt.Sprintf(getItdEndpoint, companyID, name)
	item, err := performGetITDRequest(ctx, client, endpoint)

	if err != nil {
		return "", err
	}

	data, err := item.Marshal(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func performGetITDRequest(ctx context.Context, client *client.APIClient, endpoint string) (*itd.GenericItemTypeDefinition, error) {
	resp, err := client.Get().APIPath(endpoint).Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	var itd *itd.GenericItemTypeDefinition
	if err := resp.ParseResponse(&itd); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	if itd == nil {
		return nil, fmt.Errorf("no item type definition returned in the response")
	}

	return itd, nil
}
