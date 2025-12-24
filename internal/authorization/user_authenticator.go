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

package authorization

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/oauth2"

	"github.com/mia-platform/miactl/internal/client"
)

const (
	localhost = "127.0.0.1:53535"
	appID     = "miactl"
)

type LocalServerReadyHandler func(string) error

type userAuthenticator struct {
	mutex              sync.Mutex
	userAuth           client.AuthCacheReadWriter
	client             client.Interface
	next               http.RoundTripper
	serverReadyHandler LocalServerReadyHandler
}

func (ua *userAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTrip(req, ua.next, ua.AccessToken)
}

func (ua *userAuthenticator) AccessToken() (*oauth2.Token, error) {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	jwt := ua.userAuth.ReadJWTToken()

	if jwt.Valid() {
		return jwt, nil
	}

	if refreshToken := jwt.RefreshToken; len(refreshToken) > 0 {
		return ua.refreshAuthWithToken(refreshToken)
	}

	return ua.logUser()
}

func (ua *userAuthenticator) refreshAuthWithToken(refreshToken string) (*oauth2.Token, error) {
	if token, err := ua.refreshToken(refreshToken); err == nil {
		return token, nil
	}

	return ua.logUser()
}

func (ua *userAuthenticator) logUser() (*oauth2.Token, error) {
	browserLoginConfig := &Config{
		AppID:                  appID,
		LocalServerBindAddress: []string{localhost},
		Client:                 ua.client,
		ServerReadyHandler:     ua.serverReadyHandler,
	}

	jwt, err := browserLoginConfig.GetToken(context.Background())
	if jwt != nil {
		ua.userAuth.WriteJWTToken(jwt)
	}

	return jwt, err
}

func (ua *userAuthenticator) refreshToken(token string) (*oauth2.Token, error) {
	browserLoginConfig := &Config{
		Client: ua.client,
	}

	jwt, err := browserLoginConfig.RefreshToken(context.Background(), token)
	if jwt != nil {
		ua.userAuth.WriteJWTToken(jwt)
	}
	return jwt, err
}
