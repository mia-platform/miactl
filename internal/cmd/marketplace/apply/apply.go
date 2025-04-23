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
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/files"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	applyLong = `Create or update one or more Marketplace items.

The flag -f accepts either files or directories. In case of directories, it explores them recursively.

Supported formats are JSON (.json files) and YAML (.yaml or .yml files).

The file can contain an image object with the following format:
"image": {
	"localPath": "./someImage.png"
}
The localPath can be absolute or relative to the file location.
The image will be uploaded along with the marketplace item.
Before being applied, the "image" key will be replaced with the "imageUrl" referring to the uploaded image.
You can retrieve the updated item with the "get" command.

You can also specify the "supportedByImage" in a similar fashion.

Be aware that the presence of both "image" and "imageUrl" and/or of both "supportedByImage" and "supportedByImageUrl" is illegal.`

	applyExample = `
# Apply the configuration of the file myFantasticGoTemplate.json located in the current directory to the Marketplace
miactl marketplace apply -f myFantasticGoTemplate.json

# Apply the configurations in myFantasticGoTemplate.json and myFantasticNodeTemplate.yml to the Marketplace, with relative paths
miactl marketplace apply -f ./path/to/myFantasticGoTemplate.json -f ./path/to/myFantasticNodeTemplate.yml

# Apply all the valid configuration files in the directory myFantasticGoTemplates to the Marketplace
miactl marketplace apply -f myFantasticGoTemplates`

	applyEndpointTemplate = "/api/backend/marketplace/tenants/%s/resources"

	imageAssetType = "imageAssetType"
	imageKey       = "image"
	imageURLKey    = "imageUrl"

	itemIDKey = "itemId"

	supportedByImageAssetType = "supportedByImageAssetType"
	supportedByImageKey       = "supportedByImage"
	supportedByImageURLKey    = "supportedByImageUrl"
)

var (
	errCompanyIDNotDefined = errors.New("companyID must be defined")

	errResWithoutName       = errors.New(`the required field "name" was not found in the resource`)
	errResWithoutItemID     = errors.New(`the required field "itemId" was not found in the resource`)
	errNoValidFilesProvided = errors.New("no valid files were provided")

	errResNameNotAString   = errors.New(`the field "name" must be a string`)
	errResItemIDNotAString = errors.New(`the field "itemId" must be a string`)

	errDuplicatedResIdentifier = errors.New("some resources have duplicated itemId-version tuple")
	errUnknownAssetType        = errors.New("unknown asset type")

	errUploadingImage    = errors.New("error while uploading image")
	errBuildingFilesList = errors.New("error processing files")
	errBuildingApplyReq  = errors.New("error preparing apply request")
	errProcessingImages  = errors.New("error processing images")
	errApplyingResources = errors.New("error applying items")
)

// ApplyCmd returns a new cobra command for adding or updating marketplace resources
func ApplyCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply { -f file-path }... }",
		Short:   "Create or update Marketplace items",
		Long:    applyLong,
		Example: applyExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return marketplace.ErrMissingCompanyID
			}

			outcome, err := applyItemsFromPaths(
				cmd.Context(),
				client,
				companyID,
				options.MarketplaceResourcePaths,
			)
			cobra.CheckErr(err)

			fmt.Println(outcome)

			return nil
		},
		PostRun: util.ShowDeprecatedMessage(options),
	}

	options.AddMarketplaceApplyFlags(cmd)

	return cmd
}

func applyItemsFromPaths(ctx context.Context, client *client.APIClient, companyID string, filePaths []string) (string, error) {
	resourceFilesPaths, err := buildFilePathsList(filePaths)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errBuildingFilesList, err)
	}

	applyReq, identifierToFilePathMap, err := buildApplyRequest(resourceFilesPaths)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errBuildingApplyReq, err)
	}

	for _, item := range applyReq.Resources {
		if err := processItemImages(ctx, client, companyID, item, identifierToFilePathMap); err != nil {
			return "", fmt.Errorf("%w: %s", errProcessingImages, err)
		}
	}

	outcome, err := applyMarketplaceResource(ctx, client, companyID, applyReq)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errApplyingResources, err)
	}

	return buildOutcomeSummaryAsTables(outcome), nil
}

func concatPathDirToFilePathIfRelative(basePath, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}
	itemFileDir := filepath.Dir(basePath)
	return filepath.Join(itemFileDir, filePath)
}

// processItemImages looks for image object and uploads the image when needed.
// it processes image and supportedByImage, changing the object keys with respectively imageUrl and supportedByImageUrl after the upload
func processItemImages(
	ctx context.Context,
	client *client.APIClient,
	companyID string,
	item *marketplace.Item,
	itemIDToFilePathMap map[string]string,
) error {
	processImage := func(imageObjKey, urlKey string, assetType string) error {
		localPath, err := getAndValidateImageLocalPath(item, imageObjKey, urlKey)
		if assetType != imageAssetType && assetType != supportedByImageAssetType {
			return fmt.Errorf("%w: %s", errUnknownAssetType, assetType)
		}
		if err != nil {
			return err
		}
		if localPath == "" {
			return nil
		}
		itemID := item.Get(itemIDKey).(string)
		identifier, err := buildItemIdentifier(item)
		if err != nil {
			return err
		}
		itemFilePath := itemIDToFilePathMap[identifier]
		imageFilePath := concatPathDirToFilePathIfRelative(itemFilePath, localPath)

		versionName, err := item.GetVersionName()
		if err != nil {
			return err
		}
		imageURL, err := uploadImageFileAndGetURL(
			ctx,
			client,
			companyID,
			imageFilePath,
			assetType,
			itemID,
			versionName,
		)
		if err != nil {
			return fmt.Errorf("%w: %s: %s", errUploadingImage, imageFilePath, err)
		}

		item.Del(imageObjKey)
		item.Set(urlKey, imageURL)
		return nil
	}

	if err := processImage(imageKey, imageURLKey, imageAssetType); err != nil {
		return err
	}
	err := processImage(supportedByImageKey, supportedByImageURLKey, supportedByImageAssetType)
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
	// the identifier is the concatenation of itemID and, if present, version.name
	resIdentifierToFilePath := map[string]string{}
	for _, filePath := range pathList {
		marketplaceItem := &marketplace.Item{}
		if err := files.ReadFile(filePath, marketplaceItem); err != nil {
			if errors.Is(err, files.ErrUnsupportedFile) {
				continue
			}
			return nil, nil, err
		}

		if _, err := validateItemName(marketplaceItem, filePath); err != nil {
			return nil, nil, err
		}
		itemID, err := validateItemHumanReadableID(marketplaceItem, filePath)
		if err != nil {
			return nil, nil, err
		}

		resIdentifier, err := buildItemIdentifier(marketplaceItem)
		if err != nil {
			return nil, nil, err
		}

		if _, alreadyExists := resIdentifierToFilePath[resIdentifier]; alreadyExists {
			return nil, nil, fmt.Errorf("%w: %s", errDuplicatedResIdentifier, itemID)
		}

		resources = append(resources, marketplaceItem)

		resIdentifierToFilePath[resIdentifier] = filePath
	}
	if len(resources) == 0 {
		return nil, nil, errNoValidFilesProvided
	}
	return &marketplace.ApplyRequest{
		Resources: resources,
	}, resIdentifierToFilePath, nil
}

func buildItemIdentifier(item *marketplace.Item) (string, error) {
	itemID, ok := item.Get(itemIDKey).(string)
	if !ok {
		return "", errResItemIDNotAString
	}

	versionName, err := item.GetVersionName()
	if err != nil {
		return "", err
	}

	return itemID + versionName, nil
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

func validateItemHumanReadableID(marketplaceItem *marketplace.Item, filePath string) (string, error) {
	itemID, ok := (*marketplaceItem)[itemIDKey]
	if !ok {
		return "", fmt.Errorf("%w: %s", errResWithoutItemID, filePath)
	}
	itemIDStr, ok := itemID.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s", errResItemIDNotAString, filePath)
	}
	return itemIDStr, nil
}

func applyMarketplaceResource(
	ctx context.Context,
	client *client.APIClient,
	companyID string,
	request *marketplace.ApplyRequest,
) (*marketplace.ApplyResponse, error) {
	if companyID == "" {
		return nil, errCompanyIDNotDefined
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post().
		APIPath(fmt.Sprintf(applyEndpointTemplate, companyID)).
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
