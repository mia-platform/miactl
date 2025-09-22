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

package extensions

import (
	"errors"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/files"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/spf13/cobra"
)

func ApplyCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Create or update an extension",
		Long: `Use this command to create a new extension or to update an existing one.

Extension data must be provided with a file, submitted with the -f flag; you can specificy the
extension-id either within the manifest file or via command line argument.

If an extension-id is found an updated is performed, if not instead a new extension will be created.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}

			extensionData, err := readExtensionFromFile(o.InputFilePath)
			if err != nil {
				return err
			}

			if o.EntityID != "" && extensionData.ExtensionID != "" && o.EntityID != extensionData.ExtensionID {
				return errors.New("extension id has been provided both with flags and manifest and they mismatch")
			}

			if o.EntityID != "" {
				extensionData.ExtensionID = o.EntityID
			}

			if extensionData.Type == "" {
				extensionData.Type = IFrameExtensionType
			}

			extensibilityClient := New(client)
			extensionID, err := extensibilityClient.Apply(cmd.Context(), restConfig.CompanyID, extensionData)
			cobra.CheckErr(err)
			fmt.Printf("Successfully applied extension with id %s\n", extensionID)
			return nil
		},
	}

	addExtensionIDFlag(o, cmd, "the extension id that should be edited")
	requireFilePathFlag(o, cmd)

	return cmd
}

func readExtensionFromFile(path string) (*extensibility.Extension, error) {
	extensionData := &extensibility.Extension{}
	if err := files.ReadFile(path, extensionData); err != nil {
		return nil, err
	}

	return extensionData, nil
}
