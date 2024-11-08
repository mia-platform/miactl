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

package browser

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/mia-platform/miactl/client/rest"
	"golang.org/x/oauth2"

	netutil "github.com/mia-platform/miactl/client/net/util"
)

const (
	localhost = "127.0.0.1:53535"
	appID     = "miactl"
)

type LocalServerReadyHandler func(string) error

func init() {
	if err := rest.RegisterAuthProvider(NewAuthenticator); err != nil {
		panic(fmt.Sprintf("failed to register: %s", err))
	}
}

type Authenticator struct {
	mutex              sync.Mutex
	userAuth           rest.AuthCacheReadWriter
	client             rest.Interface
	serverReadyHandler LocalServerReadyHandler
}

func NewAuthenticator(config *rest.Config, cacheReadWriter rest.AuthCacheReadWriter) rest.AuthProvider {
	client, err := rest.ClientForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	return &Authenticator{
		client:   client,
		userAuth: cacheReadWriter,
		serverReadyHandler: func(url string) error {
			if err := open(url); err != nil {
				return fmt.Errorf("could not open the browser: %w", err)
			}
			return nil
		},
	}
}

func (a *Authenticator) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &userAuthenticator{
		authenticator: a,
		next:          rt,
	}
}

func (a *Authenticator) accessToken() (*oauth2.Token, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	jwt := a.userAuth.ReadJWTToken()

	if jwt.Valid() {
		return jwt, nil
	}

	if refreshToken := jwt.RefreshToken; len(refreshToken) > 0 {
		return a.refreshAuthWithToken(refreshToken)
	}

	return a.logUser()
}

func (a *Authenticator) refreshAuthWithToken(refreshToken string) (*oauth2.Token, error) {
	if token, err := a.refreshToken(refreshToken); err == nil {
		return token, nil
	}

	return a.logUser()
}

func (a *Authenticator) logUser() (*oauth2.Token, error) {
	browserLoginConfig := &Config{
		AppID:                  appID,
		LocalServerBindAddress: []string{localhost},
		Client:                 a.client,
		ServerReadyHandler:     a.serverReadyHandler,
	}

	jwt, err := browserLoginConfig.GetToken(context.Background())
	if jwt != nil {
		a.userAuth.WriteJWTToken(jwt)
	}

	return jwt, err
}

func (a *Authenticator) refreshToken(token string) (*oauth2.Token, error) {
	browserLoginConfig := &Config{
		Client: a.client,
	}

	jwt, err := browserLoginConfig.RefreshToken(context.Background(), token)
	if jwt != nil {
		a.userAuth.WriteJWTToken(jwt)
	}
	return jwt, err
}

type userAuthenticator struct {
	authenticator *Authenticator
	next          http.RoundTripper
}

func (ua *userAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				req.Body.Close()
			}
		}()
	}

	if len(req.Header.Get("Authorization")) != 0 {
		reqBodyClosed = true
		return ua.next.RoundTrip(req)
	}

	accessToken, err := ua.authenticator.accessToken()
	if err != nil {
		return nil, err
	}

	clonedReq := netutil.CloneRequest(req)
	accessToken.SetAuthHeader(clonedReq)
	reqBodyClosed = true
	return ua.next.RoundTrip(clonedReq)
}
