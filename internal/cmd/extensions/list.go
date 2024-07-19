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

			printExtensionsList(extensions, options.Printer(clioptions.DisableWrapLines(true)))
			return nil
		},
	}
}

func printExtensionsList(extensions []*extensibility.ExtensionInfo, p printer.IPrinter) {
	p.Keys("ID", "Name", "Entry", "Destination", "Menu (id) / Category (id)", "Description")
	for _, extension := range extensions {
		p.Record(
			extension.ExtensionID,
			extension.ExtensionName,
			extension.Entry,
			extension.Destination.ID,
			menucolumn(extension),
			extension.Description,
		)
	}
	p.Print()
}

func menucolumn(extension *extensibility.ExtensionInfo) string {
	if extension.Menu.ID == "" {
		return ""
	}

	menuLabel := getTranslation(extension.Menu.LabelIntl, extensibility.En)
	if menuLabel == "" {
		return ""
	}

	menu := fmt.Sprintf("%s (%s)", menuLabel, extension.Menu.ID)

	categoryLabel := getTranslation(extension.Category.LabelIntl, extensibility.En)
	if categoryLabel != "" {
		menu += fmt.Sprintf(" / %s (%s)", categoryLabel, extension.Category.ID)
	}

	return menu
}

func getTranslation(messages extensibility.IntlMessages, defaultLang extensibility.Languages) string {
	if len(messages) == 0 {
		return ""
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
