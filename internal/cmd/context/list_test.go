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

package context

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
)

func TestPrintContexts(t *testing.T) {
	wd, _ := os.Getwd()
	testDataFolder := filepath.Join(wd, "testdata")

	testCases := map[string]struct {
		locatorPath    string
		expectedOutput string
		expectErr      bool
	}{
		"list with current context": {
			locatorPath: filepath.Join(testDataFolder, "config.yaml"),
			expectedOutput: `  context1
* context2
  context3
`,
		},
		"list without current context": {
			locatorPath: filepath.Join(testDataFolder, "missing-current.yaml"),
			expectedOutput: `  context1
  context2
  context3
`,
		},
		"error in parsing the config": {
			locatorPath:    filepath.Join(testDataFolder, "err-config"),
			expectErr:      true,
			expectedOutput: "",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = testCase.locatorPath

			buffer := bytes.NewBuffer([]byte{})
			err := printContexts(buffer, locator)
			if testCase.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedOutput, buffer.String())
		})
	}
}

func TestListContext(t *testing.T) {
	testCases := map[string]struct {
		config         api.Config
		expectedOutput []string
	}{
		"list contexts and sort names": {
			config: api.Config{
				Contexts: map[string]*api.ContextConfig{
					"context3": {},
					"context1": {},
					"context2": {},
				},
			},
			expectedOutput: []string{"context1", "context2", "context3"},
		},
		"list with nil contexts": {
			config:         api.Config{},
			expectedOutput: []string{},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			contexts := listContexts(&testCase.config)
			assert.Equal(t, testCase.expectedOutput, contexts)
		})
	}
}
