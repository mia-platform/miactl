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

	"github.com/mia-platform/miactl/internal/browser"
	"github.com/mia-platform/miactl/internal/cmd/login"
)

// nolint gosec
const tokenCachePath = ".config/miactl/cache/credentials"

type IAuth interface {
	Authenticate() (string, error)
}

type Auth struct {
	url        string
	providerID string
	browser    browser.URLOpener
	context    string
}

func (a *Auth) Authenticate() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tokenCacheAbsPath := path.Join(home, tokenCachePath)
	tokens, err := login.GetTokensFromFile(a.url, tokenCacheAbsPath)
	if err != nil {
		if !os.IsNotExist(err) && !errors.Is(err, login.ErrExpiredToken) {
			return "", err
		}
		// check for M2M login credentials
		m2mCredentialsAbsPath := path.Join(home, login.M2MCredentialsPath)
		authInfo, err := login.GetCredentialsFromFile(m2mCredentialsAbsPath, a.context)
		if err != nil {
			if !os.IsNotExist(err) && !errors.Is(err, login.ErrMissingCredentials) {
				return "", err
			}
			// if the `credentials` file does not exist, or the user did not specify m2m
			// credentials for the current context, proceed with OIDC login
			tokens, err = login.GetTokensWithOIDC(a.url, a.providerID, a.browser)
			if err != nil {
				return "", fmt.Errorf("login error: %w", err)
			}
		} else {
			// if the user specified M2M credentials for the current context,
			// proceed with M2M login
			tokens, err = login.GetTokensWithM2MLogin(a.url, authInfo)
			if err != nil {
				return "", fmt.Errorf("login error: %w", err)
			}
		}

		err = login.WriteTokensToFile(a.url, tokenCacheAbsPath, tokens)
		if err != nil {
			return "", err
		}
	}
	return tokens.AccessToken, nil
}
