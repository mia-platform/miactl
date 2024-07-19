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
	"strconv"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/spf13/cobra"
)

// ListCmd return a new cobra command for listing companies
func GetByIDCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get registered extension by id",
		Long:  "Get registered extension by id for the company.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}
			if options.EntityID == "" {
				return ErrRequiredExtensionID
			}

			extensibilityClient := New(client)
			extension, err := extensibilityClient.Get(cmd.Context(), restConfig.CompanyID, options.EntityID)
			cobra.CheckErr(err)

			printExtensionInfo(extension, options.Printer())
			return nil
		},
	}
}

func printExtensionInfo(extension *extensibility.ExtensionInfo, p printer.IPrinter) {
	p.Keys(
		"ID",
		"Name",
		"Entry",
		"Type",
		"Destination ID",
		"Description",
		"Activation Contexts",
		"Permissions",
		"Visibilities (TYPE, ID)",
		"Menu (ID, ORDER)",
		"Category (ID)",
	)

	p.Record(
		extension.ExtensionID,
		extension.ExtensionName,
		extension.Entry,
		extension.Type,
		extension.Destination.ID,
		extension.Description,
		getStringFromArray(extension.ActivationContexts, identityTransformFunc),
		getStringFromArray(extension.Permissions, identityTransformFunc),
		getStringFromArray(extension.Visibilities, func(el extensibility.Visibility) string {
			return "(" + el.ContextType + ", " + el.ContextID + ")"
		}),
		transformMenuFunc(extension.Menu),
		transformCategoryFunc(extension.Category),
	)
	p.Print()
}

func getStringFromArray[A any](array []A, transformFunc func(el A) string) string {
	s := "[ "
	for idx, el := range array {
		if idx == len(array)-1 {
			s += transformFunc(el) + " ]"
			break
		}

		s += transformFunc(el) + ", "
	}

	return s
}

func identityTransformFunc(el string) string { return el }

func transformMenuFunc(el extensibility.Menu) string {
	s := "(" + el.ID

	if el.Order != nil {
		s += fmt.Sprintf(", %s", strconv.FormatFloat(*el.Order, 'f', -1, 64))
	}
	return s + ")"
}

func transformCategoryFunc(el extensibility.Category) string {
	return "(" + el.ID + ")"
}
