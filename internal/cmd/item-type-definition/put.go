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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/mia-platform/miactl/internal/files"
	itd "github.com/mia-platform/miactl/internal/resources/item-type-definition"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

var (
	ErrInvalidFilePath        = errors.New("invalid file path")
	ErrPathIsFolder           = errors.New("path must be a file path, not a folder")
	ErrFileFormatNotSupported = errors.New("file format not supported, supported formats are: .json, .yaml, .yml")
	ErrResWithoutName         = errors.New("name field is required")
	ErrResNameNotAString      = errors.New("name field must be string")
	ErrPuttingResources       = errors.New("cannot save the item type definition")
)

const (
	putItdEndpoint = "/api/tenants/%s/marketplace/item-type-definitions/"

	cmdPutLongDescription = ` Create or update an Item Type Definition. It works with Mia-Platform Console v14.1.0 or later.

  You need to specify the flag --file or -f that accepts a file and companyId.

  Supported formats are JSON (.json files) and YAML (.yaml or .yml files). The company-id flag can be omitted if it is already set in the context.`

	putExample = `
  # Create the item type definition in file myFantasticPluginDefinition.json located in the current directory
  miactl catalog put --file myFantasticPluginDefinition.json

  # Create the item type definition in file myFantasticPluginDefinition.json, with relative path
  miactl catalog put --file ./path/to/myFantasticPluginDefinition.json`

	putCmdUse = "put { --file file-path }"
)

func PutCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     putCmdUse,
		Short:   "Create or update an Item Type Definition",
		Long:    cmdPutLongDescription,
		Example: putExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			canUseNewAPI, versionError := util.VersionCheck(cmd.Context(), client, 14, 1)
			if !canUseNewAPI || versionError != nil {
				return itd.ErrUnsupportedCompanyVersion
			}

			companyID := restConfig.CompanyID
			if len(companyID) == 0 {
				return itd.ErrMissingCompanyID
			}

			outcome, err := putItemFromPath(
				cmd.Context(),
				client,
				companyID,
				options.ItemTypeDefinitionFilePath,
				options.OutputFormat,
			)
			cobra.CheckErr(err)

			fmt.Println(outcome)

			return nil
		},
	}

	options.AddItemTypeDefinitionFileFlag(cmd)

	return cmd
}

func putItemFromPath(ctx context.Context, client *client.APIClient, companyID string, filePath string, outputFormat string) (string, error) {
	_, err := checkFilePath(filePath)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, err)
	}

	outcome, err := putItemTypeDefinition(ctx, client, companyID, filePath)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrPuttingResources, err)
	}

	data, err := outcome.Marshal(outputFormat)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func checkFilePath(rootPath string) (string, error) {
	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return ErrPathIsFolder
		}
		extension := filepath.Ext(path)
		if extension == encoding.YmlExtension || extension == encoding.YamlExtension ||
			extension == encoding.JSONExtension {
			return nil
		}
		return ErrFileFormatNotSupported
	})
	if err != nil {
		return "", err
	}
	return rootPath, nil
}

func putItemTypeDefinition(
	ctx context.Context,
	client *client.APIClient,
	companyID string,
	filePath string,
) (*itd.GenericItemTypeDefinition, error) {
	if companyID == "" {
		return nil, itd.ErrMissingCompanyID
	}

	itemTypeDefinition := &itd.GenericItemTypeDefinition{}
	if err := files.ReadFile(filePath, itemTypeDefinition); err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(itemTypeDefinition)
	if err != nil {
		return nil, err
	}

	resp, err := client.Put().
		APIPath(fmt.Sprintf(putItdEndpoint, companyID)).
		Body(bodyBytes).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if err := resp.Error(); err != nil {
		return nil, err
	}

	putResponse := &itd.GenericItemTypeDefinition{}
	err = resp.ParseResponse(putResponse)
	if err != nil {
		return nil, err
	}

	return putResponse, nil
}
