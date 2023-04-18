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

package httphandler

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/stretchr/testify/require"
)

const (
	testURLSha        = "b64868a6476817bde1123f534334c2ce78891fcad65c06667acbfdb9007b5dff"
	testTokens        = `{"accessToken":"test_token","refreshToken":"","expiresAt":9999999999}`
	testExpiredTokens = `{"accessToken":"test_token","refreshToken":"","expiresAt":0}`
	invalidJSON       = `invalid_json`
	validCredentials  = `context1:
  type: basic
  basic:
    clientId: id
    clientSecret: secret
context2:
  type: jwt
  jwt:
    key: 123abc
    algo: 123
default:
  type: basic
  basic:
    clientId: default
    clientSecret: default`
)

var (
	testDirPath string
)

func TestReadCredentials(t *testing.T) {
	testDirPath = t.TempDir()
	filePath := path.Join(testDirPath, "credentials")
	err := os.WriteFile(filePath, []byte(validCredentials), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	expectedCredentials := map[string]login.M2MAuthInfo{
		"context1": {
			AuthType: "basic",
			BasicAuth: login.BasicAuthCredentials{
				ClientID:     "id",
				ClientSecret: "secret",
			},
		},
		"context2": {
			AuthType: "jwt",
			JWTAuth: login.JWTCredentials{
				Key:  "123abc",
				Algo: "123",
			},
		},
		"default": {
			AuthType: "basic",
			BasicAuth: login.BasicAuthCredentials{
				ClientID:     "default",
				ClientSecret: "default",
			},
		},
	}

	credentials, err := readCredentials(filePath)
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
		expectedCredentials login.M2MAuthInfo
		expectedError       error
	}{
		{
			name:        "existing context",
			fileContent: validCredentials,
			context:     "context1",
			expectedCredentials: login.M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: login.BasicAuthCredentials{
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
			expectedCredentials: login.M2MAuthInfo{
				AuthType: "basic",
				BasicAuth: login.BasicAuthCredentials{
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
			expectedCredentials: login.M2MAuthInfo{},
			expectedError:       errMissingCredentials,
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := os.WriteFile(filePath, []byte(tc.fileContent), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		credentials, err := getCredentialsFromFile(filePath, tc.context)
		require.Equal(t, tc.expectedError, err)
		require.EqualValues(t, tc.expectedCredentials, credentials)
	}

}

func TestGetURLSha(t *testing.T) {
	sha := getURLSha(testBaseURL)
	require.Equal(t, testURLSha, sha)
}

func TestGetTokensFromFile(t *testing.T) {
	testDirPath = t.TempDir()
	testFilePath := path.Join(testDirPath, testURLSha)

	// valid JSON
	err := os.WriteFile(testFilePath, []byte(testTokens), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := getTokensFromFile(testBaseURL, testDirPath)
	require.NoError(t, err)
	var expectedTokens login.Tokens
	err = json.Unmarshal([]byte(testTokens), &expectedTokens)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, expectedTokens, *tokens)

	// invalid JSON
	err = os.WriteFile(testFilePath, []byte(invalidJSON), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getTokensFromFile(testBaseURL, testDirPath)
	require.ErrorContains(t, err, "invalid character")

	// expired token
	err = os.WriteFile(testFilePath, []byte(testExpiredTokens), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_, err = getTokensFromFile(testBaseURL, testDirPath)
	require.ErrorIs(t, err, errExpiredToken)
}

func TestWriteTokensToFile(t *testing.T) {
	testDirPath = t.TempDir()
	testFilePath := path.Join(testDirPath, testURLSha)

	var tokens = &login.Tokens{
		AccessToken:  "test_token",
		RefreshToken: "",
		ExpiresAt:    9999999999,
	}

	err := writeTokensToFile(testBaseURL, testDirPath, tokens)
	require.NoError(t, err)
	require.FileExists(t, testFilePath)
	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, testTokens, string(fileContent))
}
