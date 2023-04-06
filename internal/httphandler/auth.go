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
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mitchellh/go-homedir"
)

const credentialsPath = ".config/miactl/credentials"

type IAuth interface {
	Authenticate() (string, error)
}

type Auth struct {
	url        string
	providerID string
	browser    login.BrowserI
}

func (a *Auth) Authenticate() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	credentialsAbsPath := path.Join(home, credentialsPath)
	tokens, err := getTokensFromFile(a.url, credentialsAbsPath)
	if err != nil {
		if !os.IsNotExist(err) && !errors.Is(err, errExpiredToken) {
			return "", err
		} else {
			tokens, err = login.GetTokensWithOIDC(a.url, a.providerID, a.browser)
			if err != nil {
				return "", fmt.Errorf("login error: %w", err)
			}

			err = writeTokensToFile(a.url, credentialsAbsPath, tokens)
			if err != nil {
				return "", err
			}
		}
	}
	return tokens.AccessToken, nil
}
