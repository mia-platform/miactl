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
	"encoding/json"
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
	applyEndpoint = "/api/backend/marketplace/tenants/%s/resources"

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
			resources, err := buildResourcesList(resourceFilesPaths)
			cobra.CheckErr(err)
			return applyMarketplaceResource(client, resources)
		},
	}

	options.AddMarketplaceApplyFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("file", "directory")

	return cmd
}

func applyMarketplaceResource(client *client.APIClient, resources []string) error {
	return errors.New("not implemented")
}

func buildPathsListFromDir(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	filePaths := []string{}
	for _, f := range files {
		switch filepath.Ext(f.Name()) {
		case encoding.YamlExtension:
			fallthrough
		case encoding.YmlExtension:
			fallthrough
		case encoding.JsonExtension:
			filePaths = append(filePaths, f.Name())
		default:
			fmt.Printf("warning: file %s ignored because it is neither a JSON nor a YAML file\n", f.Name())
		}
	}
	return filePaths, nil
}

func buildResourcesList(pathList []string) ([]string, error) {
	resources := []string{}
	for _, path := range pathList {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		switch filepath.Ext(path) {
		case encoding.YamlExtension:
			fallthrough
		case encoding.YmlExtension:
			object := map[string]interface{}{}
			err := encoding.UnmarshalData(content, encoding.YAML, object) 
			if err != nil {
				return nil, err
			}
			jsonFormatted, err := encoding.MarshalData(object, encoding.JSON, encoding.MarshalOptions{}) 
			if err != nil {
				return nil, err
			}
			resources = append(resources, string(jsonFormatted))
		case encoding.JsonExtension:
			if json.Valid(content) {
				resources = append(resources, string(content))
			}	
		default:
			return nil, fmt.Errorf("Unsupported format\n")
		} 
		
	}
	return resources, nil
}

func parseResponse(response ApplyResponse) {}
