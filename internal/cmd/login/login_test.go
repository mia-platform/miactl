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

func TestNewLoginCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewLoginCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestReadCredentials(t *testing.T) {
	testDirPath = t.TempDir()
	filePath := path.Join(testDirPath, "credentials")
	err := os.WriteFile(filePath, []byte(validCredentials), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	expectedCredentials := map[string]M2MAuthInfo{
		"context1": {
			AuthType: "basic",
			BasicAuth: BasicAuthCredentials{
				ClientID:     "id",
				ClientSecret: "secret",
			},
		},
		"context2": {
			AuthType: "jwt",
			JWTAuth: JWTCredentials{
				Key:  "123abc",
				Algo: "123",
			},
		},
		"default": {
			AuthType: "basic",
			BasicAuth: BasicAuthCredentials{
				ClientID:     "default",
				ClientSecret: "default",
			},
		},
	}

	credentials, err := ReadCredentials(filePath)
	require.NoError(t, err)
	require.EqualValues(t, expectedCredentials, credentials)
}

func TestGetCredentialsFromFile(t *testing.T) {
	testDirPath = t.TempDir()
	filePath := path.Join(testDirPath, "credentials")

	testCases := []struct {
		name                string
		fileContent         string
		context             string
		expectedCredentials M2MAuthInfo
		expectedError       error
	}{
		{
			name:        "existing context",
			fileContent: validCredentials,
			context:     "context1",
			expectedCredentials: M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: BasicAuthCredentials{
					ClientID:     "id",
					ClientSecret: "secret",
				},
			},
			expectedError: nil,
		},
		{
			name:        "default context",
			fileContent: validCredentials,
			context:     "missing",
			expectedCredentials: M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: BasicAuthCredentials{
					ClientID:     "default",
					ClientSecret: "default",
				},
			},
			expectedError: nil,
		},
		{
			name:                "missing credentials",
			fileContent:         "",
			context:             "any",
			expectedCredentials: M2MAuthInfo{},
			expectedError:       ErrMissingCredentials,
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := os.WriteFile(filePath, []byte(tc.fileContent), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		credentials, err := GetCredentialsFromFile(filePath, tc.context)
		require.Equal(t, tc.expectedError, err)
		require.EqualValues(t, tc.expectedCredentials, credentials)
	}
}
