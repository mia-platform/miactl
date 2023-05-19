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

	"github.com/mia-platform/miactl/internal/client"
)

const (
	localhost = "127.0.0.1:53535"
	appID     = "miactl"
)

type userAuthenticator struct {
	mutex    sync.Mutex
	userAuth client.AuthCacheReadWriter
	client   client.Interface
	next     http.RoundTripper
}

func (ua *userAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return ua.next.RoundTrip(req)
	}

	return ua.next.RoundTrip(req)
}

func (ua *userAuthenticator) AccessToken() (string, error) {
	ua.mutex.Lock()
	defer ua.mutex.Unlock()

	jwt := ua.userAuth.ReadJWTToken()

	if jwt.Valid() {
		return jwt.AccessToken, nil
	}

	if refreshToken := jwt.RefreshToken; len(refreshToken) > 0 {
		return ua.refreshAuthWithToken(refreshToken)
	}

	return ua.logUser()
}

func (ua *userAuthenticator) refreshAuthWithToken(_ string) (string, error) {
	// TODO: implement refresh logic
	return ua.logUser()
}

func (ua *userAuthenticator) logUser() (string, error) {
	browserLoginConfig := Config{
		AppID:                  appID,
		LocalServerBindAddress: []string{localhost},
		Client:                 ua.client,
	}

	jwt, err := browserLoginConfig.GetToken(context.Background())
	if err != nil {
		return "", nil
	}

	ua.userAuth.WriteJWTToken(jwt)
	return jwt.AccessToken, nil
}
