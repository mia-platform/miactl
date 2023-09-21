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
	"os"
	"path/filepath"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/filesutil"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	applyLong = `Create or update one or more Marketplace items.

You can either specify:
    - one or more files, with the flag -f
    - one or more directories, with the flag -d

Supported formats are JSON (.json files) and YAML (.yaml or .yml files).`

	applyExample = `
# Apply the configuration of the file myFantasticGoTemplate.json located in the current directory to the Marketplace
miactl marketplace apply -f myFantasticGoTemplate.json

# Apply the configurations in myFantasticGoTemplate.json and myFantasticNodeTemplate.yml to the Marketplace, with relative paths
miactl marketplace apply -f ./path/to/myFantasticGoTemplate.json -f ./path/to/myFantasticNodeTemplate.yml

# Apply all the valid configuration files in the directory myFantasticGoTemplates to the Marketplace
miactl marketplace apply -d myFantasticGoTemplates`
)

// ApplyCmd returns a new cobra command for adding or updating marketplace resources
func ApplyCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply { { -f file-path }... | { -d directory-path }... }",
		Short:   "Create or update Marketplace items",
		Long:    applyLong,
		Example: applyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.MarketplaceResourcesDirPath == "" && len(options.MarketplaceResourceFilePaths) == 0 {
				return errors.New(`at least one of  "directory" or "file" must be set`)
			}

			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			var resourceFilesPaths []string
			if len(options.MarketplaceResourceFilePaths) > 0 {
				resourceFilesPaths = options.MarketplaceResourceFilePaths
			}
			if options.MarketplaceResourcesDirPath != "" {
				resourceFilesPaths, err = buildPathsListFromDir(options.MarketplaceResourcesDirPath)
				cobra.CheckErr(err)
			}
			applyReq, err := buildApplyRequest(resourceFilesPaths)
			cobra.CheckErr(err)
			outcome, err := applyMarketplaceResource(cmd.Context(), client, restConfig.CompanyID, applyReq)
			cobra.CheckErr(err)

			fmt.Println(buildOutcomeSummaryAsTables(outcome))

			return nil
		},
	}

	options.AddMarketplaceApplyFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("file", "directory")

	return cmd
}

const applyEndpoint = "/api/backend/marketplace/tenants/%s/resources"

var (
	errResWithoutName       = errors.New(`the required field "name" was not found in the resource`)
	errNoValidFilesProvided = errors.New("no valid files were provided, see errors above")
	errDuplicatedResName    = errors.New("some resources have duplicated name field")
	errResNameNotAString    = errors.New(`the field "name" must be a string`)
	errInvalidExtension     = errors.New("file has an invalid extension. Valid extensions are `.json`, `.yaml` and `.yml`")
)

func buildPathsListFromDir(dirPath string) ([]string, error) {
	filesPaths, err := filesutil.ListFilesInDirRecursive(dirPath)
	if err != nil {
		return nil, err
	}
	filePaths := []string{}
	for _, path := range filesPaths {
		switch filepath.Ext(path) {
		case encoding.YamlExtension, encoding.YmlExtension, encoding.JSONExtension:
			filePaths = append(filePaths, path)
		default:
			return nil, fmt.Errorf("%w: %s", errInvalidExtension, path)
		}
	}
	return filePaths, nil
}

func buildApplyRequest(pathList []string) (*ApplyRequest, error) {
	resources := []Item{}
	resNameToFilePath := map[string]string{}
	for _, filePath := range pathList {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var fileEncoding string
		switch filepath.Ext(filePath) {
		case encoding.YamlExtension, encoding.YmlExtension:
			fileEncoding = YAML
		case encoding.JSONExtension:
			fileEncoding = JSON
		default:
			return nil, fmt.Errorf("%w: %s", errInvalidExtension, filePath)
		}
		marketplaceItem := &Item{}
		err = encoding.UnmarshalData(content, fileEncoding, marketplaceItem)
		if err != nil {
			return nil, fmt.Errorf("errors in file %s: %w", filePath, err)
		}
		resName, err := retrieveAndValidateResName(*marketplaceItem, resNameToFilePath, filePath)
		if err != nil {
			return nil, err
		}
		resources = append(resources, *marketplaceItem)
		resNameToFilePath[resName] = filePath
	}
	if len(resources) == 0 {
		return nil, errNoValidFilesProvided
	}
	return &ApplyRequest{
		Resources: resources,
	}, nil
}

func retrieveAndValidateResName(res Item, resNameToFilePath map[string]string, filePath string) (string, error) {
	resName, ok := res["name"]
	if !ok {
		return "", fmt.Errorf("%w: %s", errResWithoutName, filePath)
	}
	resNameStr, ok := resName.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s", errResNameNotAString, filePath)
	}
	if _, ok := resNameToFilePath[resNameStr]; ok {
		return "", fmt.Errorf("%w: %s", errDuplicatedResName, resName)
	}

	return resNameStr, nil
}

func applyMarketplaceResource(ctx context.Context, client *client.APIClient, companyID string, request *ApplyRequest) (*ApplyResponse, error) {
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

	if err != nil {
		return nil, err
	}

	applyResponse := &ApplyResponse{}

	err = resp.ParseResponse(applyResponse)
	if err != nil {
		return nil, err
	}

	return applyResponse, nil
}

func buildTable(headers []string, items []ApplyResponseItem, columnTransform func(item ApplyResponseItem) []string) string {
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

func buildSuccessTable(items []ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Status"}
	columnTransform := func(item ApplyResponseItem) []string {
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

func buildFailureTable(items []ApplyResponseItem) string {
	headers := []string{"Item ID", "Name", "Validation Errors"}
	columnTransform := func(item ApplyResponseItem) []string {
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

func buildOutcomeSummaryAsTables(outcome *ApplyResponse) string {
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

func separateSuccessAndFailures(items []ApplyResponseItem) ([]ApplyResponseItem, []ApplyResponseItem) {
	var successfulItems, failedItems []ApplyResponseItem
	for _, item := range items {
		if item.Done {
			successfulItems = append(successfulItems, item)
		} else {
			failedItems = append(failedItems, item)
		}
	}
	return successfulItems, failedItems
}
