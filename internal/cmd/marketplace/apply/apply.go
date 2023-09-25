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
	"path"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
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

	// applyEndpoint has to be `Sprintf`ed with the companyID
	applyEndpoint = "/api/backend/marketplace/tenants/%s/resources"

	imageKey    = "image"
	imageURLKey = "imageUrl"

	supportedByImageKey    = "supportedByImage"
	supportedByImageURLKey = "supportedByImageUrl"
)

var (
	errCompanyIDNotDefined = errors.New("companyID must be defined")

	errResWithoutName       = errors.New(`the required field "name" was not found in the resource`)
	errNoValidFilesProvided = errors.New("no valid files were provided, see errors above")

	errResNameNotAString = errors.New(`the field "name" must be a string`)
	errInvalidExtension  = errors.New("file has an invalid extension. Valid extensions are `.json`, `.yaml` and `.yml`")
	errDuplicatedResName = errors.New("some resources have duplicated name field")
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
				restConfig.CompanyID,
				options.MarketplaceResourcePaths,
			)
			cobra.CheckErr(err)

			fmt.Println(outcome)

			return nil
		},
	}

	options.AddMarketplaceApplyFlags(cmd)

	return cmd
}

func applyItemsFromPaths(ctx context.Context, client *client.APIClient, companyID string, filePaths []string) (string, error) {
	resourceFilesPaths, err := buildFilePathsList(filePaths)
	if err != nil {
		return "", err
	}
	applyReq, itemNameToFilePath, err := buildApplyRequest(resourceFilesPaths)
	if err != nil {
		return "", err
	}

	for _, item := range applyReq.Resources {
		if err := processItemImages(ctx, client, companyID, item, itemNameToFilePath); err != nil {
			return "", err
		}
	}

	outcome, err := applyMarketplaceResource(ctx, client, companyID, applyReq)
	if err != nil {
		return "", err
	}

	return buildOutcomeSummaryAsTables(outcome), nil
}

// processItemImages looks for image object and uploads the image when needed.
// it processes image and supportedByImage, changing the object keys with respectively imageUrl and supportedByImageUrl after the upload
func processItemImages(ctx context.Context, client *client.APIClient, companyID string, item *marketplace.Item, itemNameToFilePath map[string]string) error {
	processImage := func(objKey, urlKey string) error {
		localPath, err := getAndValidateImageLocalPath(item, objKey, urlKey)
		if err != nil {
			return err
		}
		if localPath == "" {
			return nil
		}
		itemName := (*item)["name"].(string)
		itemFilePath := itemNameToFilePath[itemName]
		itemFileDir := filepath.Dir(itemFilePath)
		imageFilePath := path.Join(itemFileDir, localPath)

		imageURL, err := uploadImageFileAndGetURL(ctx, client, companyID, imageFilePath)
		if err != nil {
			return err
		}

		item.Del(objKey)
		item.Set(urlKey, imageURL)
		return nil
	}

	if err := processImage(imageKey, imageURLKey); err != nil {
		return err
	}
	err := processImage(supportedByImageKey, supportedByImageURLKey)
	return err
}

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

func buildApplyRequest(pathList []string) (*marketplace.ApplyRequest, map[string]string, error) {
	resources := []*marketplace.Item{}
	resNameToFilePath := map[string]string{}
	for _, filePath := range pathList {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, nil, err
		}
		var fileEncoding string
		switch filepath.Ext(filePath) {
		case encoding.YamlExtension, encoding.YmlExtension:
			fileEncoding = encoding.YAML
		case encoding.JSONExtension:
			fileEncoding = encoding.JSON
		default:
			return nil, nil, fmt.Errorf("%w: %s", errInvalidExtension, filePath)
		}

		marketplaceItem := &marketplace.Item{}
		err = encoding.UnmarshalData(content, fileEncoding, marketplaceItem)
		if err != nil {
			return nil, nil, fmt.Errorf("errors in file %s: %w", filePath, err)
		}

		itemNameStr, err := validateItemName(marketplaceItem, filePath)
		if err != nil {
			return nil, nil, err
		}
		if _, alreadyExists := resNameToFilePath[itemNameStr]; alreadyExists {
			return nil, nil, fmt.Errorf("%w: %s", errDuplicatedResName, itemNameStr)
		}

		resources = append(resources, marketplaceItem)
		resNameToFilePath[itemNameStr] = filePath
	}
	if len(resources) == 0 {
		return nil, nil, errNoValidFilesProvided
	}
	return &marketplace.ApplyRequest{
		Resources: resources,
	}, resNameToFilePath, nil
}

func validateItemName(marketplaceItem *marketplace.Item, filePath string) (string, error) {
	itemName, ok := (*marketplaceItem)["name"]
	if !ok {
		return "", fmt.Errorf("%w: %s", errResWithoutName, filePath)
	}
	itemNameStr, ok := itemName.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s", errResNameNotAString, filePath)
	}
	return itemNameStr, nil
}

func applyMarketplaceResource(ctx context.Context, client *client.APIClient, companyID string, request *marketplace.ApplyRequest) (*marketplace.ApplyResponse, error) {
	if companyID == "" {
		return nil, errCompanyIDNotDefined
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
