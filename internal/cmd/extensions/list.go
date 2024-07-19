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
	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/spf13/cobra"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered extensions",
		Long:  "List registered extensions for the company.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}

			extensibilityClient := New(client)
			extensions, err := extensibilityClient.List(cmd.Context(), restConfig.CompanyID)
			cobra.CheckErr(err)

			printExtensionsList(extensions, options.Printer())
			return nil
		},
	}
}

func printExtensionsList(extensions []*extensibility.Extension, p printer.IPrinter) {
	p.Keys("ID", "Name", "Description")
	for _, extension := range extensions {
		p.Record(
			extension.ExtensionID,
			extension.Name,
			extension.Description,
		)
	}
	p.Print()
}
