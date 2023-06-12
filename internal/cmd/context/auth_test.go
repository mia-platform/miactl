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

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAuth(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(wd, "testdata")
	testCases := map[string]struct {
		configPath     string
		authName       string
		options        *clioptions.CLIOptions
		expectOverride bool
	}{
		"empty file": {
			authName:   "credential",
			configPath: filepath.Join(t.TempDir(), "empty"),
			options: &clioptions.CLIOptions{
				BasicClientID:     "id",
				BasicClientSecret: "secret",
			},
		},
		"existing file": {
			authName:   "credentialTest",
			configPath: copyFile(t, filepath.Join(testdata, "auth.yaml")),
			options: &clioptions.CLIOptions{
				BasicClientID:     "id",
				BasicClientSecret: "secret",
			},
		},
		"merge auth": {
			authName:   "credential1",
			configPath: copyFile(t, filepath.Join(testdata, "auth.yaml")),
			options: &clioptions.CLIOptions{
				BasicClientSecret: "secret",
			},
			expectOverride: true,
		},
		"config with only contexts": {
			authName:   "credentialTest",
			configPath: copyFile(t, filepath.Join(testdata, "config.yaml")),
			options: &clioptions.CLIOptions{
				BasicClientID:     "id",
				BasicClientSecret: "secret",
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			tempFile := testCase.configPath

			testCase.options.MiactlConfig = tempFile
			override, err := setAuth(testCase.authName, testCase.options)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectOverride, override)
		})
	}
}
