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
	testURLSha  = "b64868a6476817bde1123f534334c2ce78891fcad65c06667acbfdb9007b5dff"
	testTokens  = `{"accessToken":"test_token"}`
	invalidJSON = `invalid_json`
)

var (
	testDirPath string
)

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
}

func TestWriteTokensToFile(t *testing.T) {

}
