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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/encoding"
	"github.com/spf13/cobra"
)

const (
	applyLong = `Create or update one or more Marketplace items.
	You can either specify one or more files or one or more directories, respectively with the flags -f and -k.`
	applyExample = `
	# Apply the configuration in myFantasticGoTemplate.json to the Marketplace
	miactl marketplace apply -f myFantasticGoTemplate.json

	# Apply the configurations in myFantasticGoTemplate.json and myFantasticNodeTemplate.json to the Marketplace
	miactl marketplace apply -f myFantasticGoTemplate.json -f myFantasticNodeTemplate.json

	# Apply all the configurations in the folder myFantasticGoTemplates to the Marketplace
	miactl marketplace apply -k myFantasticGoTemplates`
)

// ApplyCmd returns a new cobra command for adding or updating marketplace resources
func ApplyCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply { { -f file }... | { -k directory }... }",
		Short:   "Create or update Marketplace items",
		Long:    applyLong,
		Example: applyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return applyMarketplaceResource(client, applyReq)
		},
	}

	options.AddMarketplaceApplyFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("file", "directory")

	return cmd
}

const (
	applyEndpoint           = "/api/backend/marketplace/tenants/%s/resources"
	invalidExtensionWarning = "warning: file %s was ignored because it has not a recognized extension. Valid extensions are `.json`, `.yaml` and `.yml`\n"

	errParsingFile = "error parsing file: %s"
)

var errNoValidFilesProvided = errors.New("no valid files were provided.")

func applyMarketplaceResource(client *client.APIClient, request *ApplyRequest) error {
	return errors.New("not implemented")
}

// listFiles recursively lists file in the given directory path
func listFiles(rootPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func buildPathsListFromDir(dirPath string) ([]string, error) {
	filesPaths, err := listFiles(dirPath)
	if err != nil {
		return nil, err
	}
	filePaths := []string{}
	for _, path := range filesPaths {
		switch filepath.Ext(path) {
		case encoding.YamlExtension:
			fallthrough
		case encoding.YmlExtension:
			fallthrough
		case encoding.JsonExtension:
			filePaths = append(filePaths, path)
		default:
			fmt.Printf(invalidExtensionWarning, path)
		}
	}
	return filePaths, nil
}

func buildApplyRequest(pathList []string) (*ApplyRequest, error) {
	resources := []*map[string]interface{}{}
	for _, path := range pathList {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %w", err)
		}
		object := &map[string]interface{}{}
		var fileEncoding string
		switch filepath.Ext(path) {
		case encoding.YamlExtension:
			fallthrough
		case encoding.YmlExtension:
			fileEncoding = YAML
		case encoding.JsonExtension:
			fileEncoding = JSON
		default:
			fmt.Printf(invalidExtensionWarning, path)
			continue
		}
		err = encoding.UnmarshalData(content, fileEncoding, object)
		if err != nil {
			return nil, fmt.Errorf("error parsing file %s: %w", path, err)
		}
		resources = append(resources, object)
	}
	if len(resources) == 0 {
		return nil, errNoValidFilesProvided
	}
	return &ApplyRequest{
		Resources: resources,
	}, nil
}

func parseResponse(response ApplyResponse) {}
