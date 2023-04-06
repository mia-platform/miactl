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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/mia-platform/miactl/internal/cmd/login"
)

var errExpiredToken = errors.New("the token has expired")

func getTokensFromFile(url, credentialsPath string) (*login.Tokens, error) {
	sha := getURLSha(url)

	filePath := path.Join(credentialsPath, sha)

	tokenBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tokens login.Tokens
	err = json.Unmarshal(tokenBytes, &tokens)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if now.After(time.Unix(tokens.ExpiresAt, 0)) {
		return nil, errExpiredToken
	}

	return &tokens, nil
}

func writeTokensToFile(url, credentialsPath string, tokens *login.Tokens) error {
	sha := getURLSha(url)

	filePath := path.Join(credentialsPath, sha)
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	credentials := []byte(tokenJSON)

	err = os.WriteFile(filePath, credentials, os.ModePerm)
	return err
}

func getURLSha(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	bs := hasher.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
