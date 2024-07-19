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

func TestGetByIDCommandBuilder(t *testing.T) {
	opts := clioptions.NewCLIOptions()
	cmd := GetByIDCmd(opts)
	require.NotNil(t, cmd)
}

func TestPrintExtensionInfo(t *testing.T) {
	data := &extensibility.ExtensionInfo{
		ExtensionID:        "mocked-id",
		ExtensionName:      "mocked-name",
		Entry:              "http://example.com/",
		Type:               "iframe",
		Destination:        extensibility.DestinationArea{ID: "project"},
		Description:        "some description",
		ActivationContexts: []string{"project", "company"},
		Permissions:        []string{"perm1", "perm2"},
		Visibilities:       []extensibility.Visibility{{ContextType: "project", ContextID: "prjId"}},
		Menu: extensibility.Menu{
			ID: "routeId",
			LabelIntl: map[string]string{
				"en": "some label",
				"it": "qualche etichetta",
			},
		},
		Category: extensibility.Category{
			ID: "some-category",
		},
	}

	str := &strings.Builder{}
	printExtensionInfo(
		data,
		printer.NewTablePrinter(printer.TablePrinterOptions{}, str),
	)

	expectedTokens := []string{
		"ID", "NAME", "ENTRY", "TYPE", "DESTINATION ID", "DESCRIPTION", "ACTIVATION CONTEXTS", "PERMISSIONS", "VISIBILITIES (TYPE, ID)", "MENU (ID, ORDER)", "CATEGORY (ID)",
		"mocked-id", "mocked-name", "http://example.com/", "iframe", "project", "some description", "[ project, company ]", "[ perm1, perm2 ]", "[ (project, prjId) ]", "(routeId)", "(some-category)",
	}

	for _, expected := range expectedTokens {
		require.Contains(t, str.String(), expected)
	}
}
