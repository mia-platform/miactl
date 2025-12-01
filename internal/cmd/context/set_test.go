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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/clioptions"
)

func TestSetContext(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(wd, "testdata")
	testCases := map[string]struct {
		configPath     string
		contextName    string
		options        *clioptions.CLIOptions
		expectOverride bool
	}{
		"empty file": {
			contextName: "context",
			configPath:  filepath.Join(t.TempDir(), "empty"),
			options: &clioptions.CLIOptions{
				Endpoint: "example.com",
			},
		},
		"existing file": {
			contextName: "contextTest",
			configPath:  copyFile(t, filepath.Join(testdata, "config.yaml")),
			options: &clioptions.CLIOptions{
				Endpoint: "example.com",
			},
		},
		"merge context": {
			contextName: "context1",
			configPath:  copyFile(t, filepath.Join(testdata, "config.yaml")),
			options: &clioptions.CLIOptions{
				CompanyID: "company",
			},
			expectOverride: true,
		},
		"config with only auth": {
			contextName: "contextTest",
			configPath:  copyFile(t, filepath.Join(testdata, "auth.yaml")),
			options: &clioptions.CLIOptions{
				CompanyID: "company",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			tempFile := testCase.configPath

			testCase.options.MiactlConfig = tempFile
			override, err := setContext(testCase.contextName, testCase.options)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectOverride, override)
		})
	}
}
