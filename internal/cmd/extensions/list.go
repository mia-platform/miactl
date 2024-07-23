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
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/spf13/cobra"
)

const noLabel = "NO LABEL"

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
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
			extensions, err := extensibilityClient.List(cmd.Context(), restConfig.CompanyID, options.ResolveExtensionsDetails)
			cobra.CheckErr(err)

			printExtensionsList(extensions, options.Printer(clioptions.DisableWrapLines(true)), options.ResolveExtensionsDetails)
			return nil
		},
	}

	addResolveDetailsFlag(options, cmd)
	return cmd
}

func addResolveDetailsFlag(options *clioptions.CLIOptions, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(&options.ResolveExtensionsDetails, "resolve-details", false, "Retrieve also menu and category info")
}

func printExtensionsList(extensions []*extensibility.ExtensionInfo, p printer.IPrinter, resolveDetails bool) {
	tableColumnLabel := []string{"ID", "Name", "Entry", "Destination", "Description"}
	if resolveDetails {
		tableColumnLabel = append(tableColumnLabel, "Menu (id)")
		tableColumnLabel = append(tableColumnLabel, "Category (id)")
	}
	p.Keys(tableColumnLabel...)
	for _, extension := range extensions {
		tableRow := []string{
			extension.ExtensionID,
			extension.Name,
			extension.Entry,
			extension.Destination.ID,
			extension.Description,
		}
		if resolveDetails {
			if extension.Menu == nil {
				tableRow = append(tableRow, "")
			} else {
				tableRow = append(tableRow, menucolumn(extension.Menu.ID, extension.Menu.LabelIntl))
			}

			if extension.Category == nil {
				tableRow = append(tableRow, "")
			} else {
				tableRow = append(tableRow, menucolumn(extension.Category.ID, extension.Category.LabelIntl))
			}
		}
		p.Record(tableRow...)
	}
	p.Print()
}

func menucolumn(id string, labelIntl extensibility.IntlMessages) string {
	return fmt.Sprintf("%s (%s)", getTranslation(labelIntl, extensibility.En), id)
}

func getTranslation(messages extensibility.IntlMessages, defaultLang extensibility.Languages) string {
	if len(messages) == 0 {
		return noLabel
	}

	defaultMessage, ok := messages[defaultLang]
	if !ok {
		for _, msg := range messages {
			if msg != "" {
				return msg
			}
		}
	}
	return defaultMessage
}
