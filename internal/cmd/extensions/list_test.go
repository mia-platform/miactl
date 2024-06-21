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
	data := []*extensibility.Extension{
		{
			ExtensionID: "ext-1",
			Name:        "Extension 1",
			Description: "Description 1",
		},
		{
			ExtensionID: "ext-2",
			Name:        "Extension 2",
			Description: "Description 2",
		},
	}

	str := &strings.Builder{}
	printExtensionsList(
		data,
		printer.NewTablePrinter(printer.TablePrinterOptions{}, str),
	)

	expectedTokens := []string{
		"ID", "NAME", "DESCRIPTION",
		"ext-1", "Extension 1", "Description 1",
		"ext-2", "Extension 2", "Description 2",
	}

	for _, expected := range expectedTokens {
		require.Contains(t, str.String(), expected)
	}
}
