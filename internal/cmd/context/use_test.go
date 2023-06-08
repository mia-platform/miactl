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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCurrentContext(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(wd, "testdata", "config.yaml")

	testCases := map[string]struct {
		configPath string
		newContext string
		expectErr  bool
	}{
		"change context": {
			newContext: "context3",
		},
		"wrong context": {
			newContext: "foo",
			expectErr:  true,
		},
		"same context": {
			newContext: "context2",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			tempFile := copyFile(t, testdata)
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = tempFile

			err := setCurrentContext(testCase.newContext, locator)
			switch testCase.expectErr {
			case true:
				assert.Error(t, err)
			default:
				assert.NoError(t, err)
			}
		})
	}
}

func copyFile(t *testing.T, in string) string {
	t.Helper()
	inFile, err := os.OpenFile(in, os.O_RDONLY, os.ModePerm)
	require.NoError(t, err)
	outFile, err := os.CreateTemp(t.TempDir(), "use-context")
	require.NoError(t, err)

	_, err = io.Copy(outFile, inFile)
	require.NoError(t, err)
	return outFile.Name()
}
