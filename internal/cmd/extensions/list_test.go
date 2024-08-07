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
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/extensibility"

	"github.com/stretchr/testify/require"
)

func TestListCommandBuilder(t *testing.T) {
	opts := clioptions.NewCLIOptions()
	cmd := ListCmd(opts)
	require.NotNil(t, cmd)
}

func TestPrintExtensionsList(t *testing.T) {
	t.Run("without resolve details", func(t *testing.T) {
		data := []*extensibility.ExtensionInfo{
			{
				ExtensionID: "ext-1",
				Name:        "Extension 1",
				Entry:       "https://docs.mia-platform.eu",
				Destination: extensibility.DestinationArea{ID: "d1"},
				Menu:        &extensibility.Menu{ID: "menu-id", LabelIntl: extensibility.IntlMessages{"en": "Ex 1"}},
				Category:    &extensibility.Category{ID: "cat-1", LabelIntl: extensibility.IntlMessages{"en": "Cat 1"}},
				Description: "Description 1",
			},
			{
				ExtensionID: "ext-2",
				Name:        "Extension 2",
				Entry:       "https://mia-platform.eu",
				Destination: extensibility.DestinationArea{ID: "d2"},
				Menu:        &extensibility.Menu{ID: "menu-id-2", LabelIntl: extensibility.IntlMessages{"en": "Ex 2"}},
				Description: "Description 2",
			},
		}

		str := &strings.Builder{}
		printExtensionsList(
			data,
			printer.NewTablePrinter(printer.TablePrinterOptions{}, str),
			false,
		)

		expectedTokens := []string{
			"ID", "NAME", "ENTRY", "DESTINATION", "DESCRIPTION",
			"ext-1", "Extension 1", "https://docs.mia-platform.eu", "d1", "Description 1",
			"ext-2", "Extension 2", "https://mia-platform.eu", "d2", "Description 2",
		}

		for _, expected := range expectedTokens {
			require.Contains(t, str.String(), expected)
		}
	})

	t.Run("resolving details", func(t *testing.T) {
		data := []*extensibility.ExtensionInfo{
			{
				ExtensionID: "ext-1",
				Name:        "Extension 1",
				Entry:       "https://docs.mia-platform.eu",
				Destination: extensibility.DestinationArea{ID: "d1"},
				Menu:        &extensibility.Menu{ID: "menu-id", LabelIntl: extensibility.IntlMessages{"en": "Ex 1"}},
				Category:    &extensibility.Category{ID: "cat-1", LabelIntl: extensibility.IntlMessages{"en": "Cat 1"}},
				Description: "Description 1",
			},
			{
				ExtensionID: "ext-2",
				Name:        "Extension 2",
				Entry:       "https://mia-platform.eu",
				Destination: extensibility.DestinationArea{ID: "d2"},
				Menu:        &extensibility.Menu{ID: "menu-id-2", LabelIntl: extensibility.IntlMessages{"en": "Ex 2"}},
				Description: "Description 2",
			},
			{
				ExtensionID: "ext-3",
				Name:        "Extension 3",
				Entry:       "https://docs.mia-platform.eu",
				Destination: extensibility.DestinationArea{ID: "d3"},
				Menu:        &extensibility.Menu{ID: "menu-id", LabelIntl: extensibility.IntlMessages{"en": "Ex 3"}},
				Category:    &extensibility.Category{ID: "cat-3"},
				Description: "Description 3",
			},
		}

		str := &strings.Builder{}
		printExtensionsList(
			data,
			printer.NewTablePrinter(printer.TablePrinterOptions{}, str),
			true,
		)

		expectedTokens := []string{
			"ID", "NAME", "ENTRY", "DESTINATION", "DESCRIPTION", "MENU (ID)", "CATEGORY (ID)",
			"ext-1", "Extension 1", "https://docs.mia-platform.eu", "d1", "Description 1", "Ex 1 (menu-id)", "Cat 1 (cat-1)",
			"ext-2", "Extension 2", "https://mia-platform.eu", "d2", "Description 2", "Ex 2 (menu-id-2)",
			"ext-3", "Extension 3", "https://docs.mia-platform.eu", "d3", "Description 3", "Ex 3 (menu-id)", "NO LABEL (cat-3)",
		}

		for _, expected := range expectedTokens {
			require.Contains(t, str.String(), expected)
		}
	})
}
