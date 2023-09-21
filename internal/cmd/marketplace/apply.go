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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	applyLong = `Create or update one or more Marketplace items.

The flag -f accepts either files or directories. In case of directories, it explores them recursively.

Supported formats are JSON (.json files) and YAML (.yaml or .yml files).`

	applyExample = `
# Apply the configuration of the file myFantasticGoTemplate.json located in the current directory to the Marketplace
miactl marketplace apply -f myFantasticGoTemplate.json

# Apply the configurations in myFantasticGoTemplate.json and myFantasticNodeTemplate.yml to the Marketplace, with relative paths
miactl marketplace apply -f ./path/to/myFantasticGoTemplate.json -f ./path/to/myFantasticNodeTemplate.yml

# Apply all the valid configuration files in the directory myFantasticGoTemplates to the Marketplace
miactl marketplace apply -f myFantasticGoTemplates`
)

// ApplyCmd returns a new cobra command for adding or updating marketplace resources
func ApplyCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply { -f file-path }... }",
		Short:   "Create or update Marketplace items",
		Long:    applyLong,
		Example: applyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			outcome, err := applyItemsFromPaths(
				cmd.Context(),
				client,
				restConfig,
				options.MarketplaceResourcePaths,
			)
			cobra.CheckErr(err)

			fmt.Println(outcome)

			return nil
		},
	}

	options.AddMarketplaceApplyFlags(cmd.Flags())
	cmd.MarkFlagRequired("file")

	return cmd
}

func applyItemsFromPaths(ctx context.Context, client *client.APIClient, restConfig *client.Config, filePaths []string) (string, error) {
	resourceFilesPaths, err := buildFilePathsList(filePaths)
	if err != nil {
		return "", err
	}
	applyReq, err := buildApplyRequest(resourceFilesPaths)
	if err != nil {
		return "", err
	}
	outcome, err := applyMarketplaceResource(ctx, client, restConfig.CompanyID, applyReq)
	if err != nil {
		return "", err
	}

	return buildOutcomeSummaryAsTables(outcome), nil
}

const applyEndpoint = "/api/backend/marketplace/tenants/%s/resources"

var (
	errResWithoutName       = errors.New(`the required field "name" was not found in the resource`)
	errNoValidFilesProvided = errors.New("no valid files were provided, see errors above")
	errDuplicatedResName    = errors.New("some resources have duplicated name field")
	errResNameNotAString    = errors.New(`the field "name" must be a string`)
	errInvalidExtension     = errors.New("file has an invalid extension. Valid extensions are `.json`, `.yaml` and `.yml`")
)

func buildFilePathsList(paths []string) ([]string, error) {
	filePaths := []string{}
	for _, rootPath := range paths {
		err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			switch filepath.Ext(path) {
			case encoding.YmlExtension, encoding.YamlExtension, encoding.JSONExtension:
				filePaths = append(filePaths, path)
			default:
				return fmt.Errorf("%w: %s", errInvalidExtension, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return filePaths, nil
}

func buildApplyRequest(pathList []string) (*marketplace.ApplyRequest, error) {
	resources := []marketplace.Item{}
	resNameToFilePath := map[string]string{}
	for _, filePath := range pathList {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var fileEncoding string
		switch filepath.Ext(filePath) {
		case encoding.YamlExtension, encoding.YmlExtension:
			fileEncoding = marketplace.YAMLFormat
		case encoding.JSONExtension:
			fileEncoding = marketplace.JSON
		default:
			return nil, fmt.Errorf("%w: %s", errInvalidExtension, filePath)
		}

		marketplaceItem := &marketplace.Item{}
		err = encoding.UnmarshalData(content, fileEncoding, marketplaceItem)
		if err != nil {
			return nil, fmt.Errorf("errors in file %s: %w", filePath, err)
		}

		itemName, ok := (*marketplaceItem)["name"]
		if !ok {
			return nil, fmt.Errorf("%w: %s", errResWithoutName, filePath)
		}
		itemNameStr, ok := itemName.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s", errResNameNotAString, filePath)
		}
		if _, ok := resNameToFilePath[itemNameStr]; ok {
			return nil, fmt.Errorf("%w: %s", errDuplicatedResName, itemName)
		}

		resources = append(resources, *marketplaceItem)
		resNameToFilePath[itemNameStr] = filePath
	}
	if len(resources) == 0 {
		return nil, errNoValidFilesProvided
	}
	return &marketplace.ApplyRequest{
		Resources: resources,
	}, nil
}

func applyMarketplaceResource(ctx context.Context, client *client.APIClient, companyID string, request *marketplace.ApplyRequest) (*marketplace.ApplyResponse, error) {
	if companyID == "" {
		return nil, errors.New("companyID must be defined")
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post().
		APIPath(fmt.Sprintf(applyEndpoint, companyID)).
		Body(bodyBytes).
		Do(ctx)

	if err := resp.Error(); err != nil {
		return nil, err
	}

	applyResp := &marketplace.ApplyResponse{}

	err = resp.ParseResponse(applyResp)
	if err != nil {
		return nil, err
	}

	return applyResp, nil
}

func buildTable(headers []string, items []marketplace.ApplyResponseItem, columnTransform func(item marketplace.ApplyResponseItem) []string) string {
	strBuilder := &strings.Builder{}
	table := tablewriter.NewWriter(strBuilder)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetHeader(headers)

	for _, item := range items {
		table.Append(columnTransform(item))
	}

	table.Render()
	return strBuilder.String()
}

func buildSuccessTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Status"}
	columnTransform := func(item marketplace.ApplyResponseItem) []string {
		var status string
		switch {
		case item.Inserted:
			status = "Inserted"
		case item.Updated:
			status = "Updated"
		default:
			// should never happen, but just in case:
			status = "UNKNOWN"
		}
		return []string{item.ItemID, item.Name, status}
	}

	return buildTable(headers, items, columnTransform)
}

func buildFailureTable(items []marketplace.ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Validation Errors"}
	columnTransform := func(item marketplace.ApplyResponseItem) []string {
		var validationErrorsStr string
		validationErrors := item.ValidationErrors
		for i, valErr := range validationErrors {
			validationErrorsStr += valErr.Message
			if len(validationErrors)-1 > i {
				validationErrorsStr += "\n"
			}
		}
		if validationErrorsStr == "" {
			validationErrorsStr = "-"
		}
		return []string{item.ItemID, item.Name, validationErrorsStr}
	}

	return buildTable(headers, items, columnTransform)
}

func buildOutcomeSummaryAsTables(outcome *marketplace.ApplyResponse) string {
	successfulItems, failedItems := separateSuccessAndFailures(outcome.Items)
	successfulCount := len(successfulItems)
	failedCount := len(failedItems)

	var outcomeStr string

	if successfulCount > 0 {
		outcomeStr += fmt.Sprintf("%d of %d items have been successfully applied:\n\n", successfulCount, len(outcome.Items))
		outcomeStr += buildSuccessTable(successfulItems)
	}

	if failedCount > 0 {
		if successfulCount > 0 {
			outcomeStr += fmt.Sprintln()
		}
		outcomeStr += fmt.Sprintf("%d of %d items have not been applied due to validation errors:\n\n", failedCount, len(outcome.Items))
		outcomeStr += buildFailureTable(failedItems)
	}
	return outcomeStr
}

func separateSuccessAndFailures(items []marketplace.ApplyResponseItem) ([]marketplace.ApplyResponseItem, []marketplace.ApplyResponseItem) {
	var successfulItems, failedItems []marketplace.ApplyResponseItem
	for _, item := range items {
		if item.Done {
			successfulItems = append(successfulItems, item)
		} else {
			failedItems = append(failedItems, item)
		}
	}
	return successfulItems, failedItems
}
