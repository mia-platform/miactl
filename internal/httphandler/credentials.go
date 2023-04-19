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
	"gopkg.in/yaml.v3"
)

var errExpiredToken = errors.New("the token has expired")
var errMissingCredentials = errors.New("missing credentials for current and default context")

func readCredentials(credentialsPath string) (map[string]login.M2MAuthInfo, error) {
	yamlCredentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}
	var credentialsMap map[string]login.M2MAuthInfo
	err = yaml.Unmarshal(yamlCredentials, &credentialsMap)
	return credentialsMap, err
}

func getCredentialsFromFile(credentialsPath, context string) (login.M2MAuthInfo, error) {
	credentialsMap, err := readCredentials(credentialsPath)
	if err != nil {
		return login.M2MAuthInfo{}, err
	}
	if credential, found := credentialsMap[context]; found {
		return credential, nil
	}
	if credential, found := credentialsMap["default"]; found {
		return credential, nil
	}
	return login.M2MAuthInfo{}, errMissingCredentials
}

func getTokensFromFile(url, tokenCachePath string) (*login.Tokens, error) {
	sha := getURLSha(url)

	filePath := path.Join(tokenCachePath, sha)

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

func writeTokensToFile(url, tokenCachePath string, tokens *login.Tokens) error {
	sha := getURLSha(url)

	filePath := path.Join(tokenCachePath, sha)
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	credentials := tokenJSON

	err = os.WriteFile(filePath, credentials, os.ModePerm)
	return err
}

func getURLSha(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	bs := hasher.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
