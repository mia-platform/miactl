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
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mitchellh/go-homedir"
)

func getTokensFromFile(url string) (*login.Tokens, error) {
	sha := getURLSha(url)

	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	credentialsAbsPath := path.Join(home, credentialsPath, sha)

	tokenBytes, err := os.ReadFile(credentialsAbsPath)
	if err != nil {
		return nil, err
	}

	var tokens login.Tokens
	err = json.Unmarshal(tokenBytes, &tokens)
	if err != nil {
		return nil, err
	}

	return &tokens, nil
}

func writeTokensToFile(url string, tokens *login.Tokens) error {
	sha := getURLSha(url)

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	credentialsAbsPath := path.Join(home, credentialsPath, sha)
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	credentials := []byte(tokenJSON)

	_, err = os.Stat(credentialsAbsPath)
	if os.IsNotExist(err) {
		_, err := os.Create(credentialsAbsPath)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(credentialsAbsPath, credentials, os.ModePerm)
	return err
}

func getURLSha(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	bs := hasher.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
