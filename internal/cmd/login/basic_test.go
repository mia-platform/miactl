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

package login

import (
	"os"
	"path"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/require"
)

func TestNewBasicLoginCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewBasicLoginCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestUpdateBasicCredentials(t *testing.T) {
	testDirPath = t.TempDir()
	filePath := path.Join(testDirPath, "credentials")
	err := os.WriteFile(filePath, []byte(validCredentials), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		options          clioptions.CLIOptions
		expectedAuthInfo M2MAuthInfo
		expectedError    error
	}{
		"update existing credentials": {
			options: clioptions.CLIOptions{
				BasicClientID:     "newId",
				BasicClientSecret: "newSecret",
				Context:           "context1",
			},
			expectedAuthInfo: M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: BasicAuthCredentials{
					ClientID:     "newId",
					ClientSecret: "newSecret",
				},
			},
			expectedError: nil,
		},
		"create new credentials": {
			options: clioptions.CLIOptions{
				BasicClientID:     "id",
				BasicClientSecret: "secret",
				Context:           "context3",
			},
			expectedAuthInfo: M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: BasicAuthCredentials{
					ClientID:     "id",
					ClientSecret: "secret",
				},
			},
			expectedError: nil,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			authInfo, err := updateBasicCredentials(filePath, testCase.options)
			require.ErrorIs(t, err, testCase.expectedError)
			require.EqualValues(t, testCase.expectedAuthInfo, *authInfo)
			credentialsMap, err := ReadCredentials(filePath)
			if err != nil {
				t.Fatal(err)
			}
			require.EqualValues(t, testCase.expectedAuthInfo, credentialsMap[testCase.options.Context])
		})
	}
}
