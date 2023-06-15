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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	basicAuthEndpoint = "/api/m2m/oauth/token"
)

type basicAuthenticator struct {
	mutex        sync.Mutex
	userAuth     client.AuthCacheReadWriter
	client       client.Interface
	next         http.RoundTripper
	clientID     string
	clientSecret string
}

func (ba *basicAuthenticator) RoundTrip(req *http.Request) (*http.Response, error) {
	return roundTrip(req, ba.next, ba.AccessToken)
}

func (ba *basicAuthenticator) AccessToken() (*oauth2.Token, error) {
	ba.mutex.Lock()
	defer ba.mutex.Unlock()

	jwt := ba.userAuth.ReadJWTToken()

	if jwt.Valid() {
		return jwt, nil
	}

	return ba.basicAuth()
}

func (ba *basicAuthenticator) basicAuth() (*oauth2.Token, error) {
	jwt, err := getToken(context.Background(), ba.client, ba.clientID, ba.clientSecret)
	if jwt != nil {
		ba.userAuth.WriteJWTToken(jwt)
	}

	return jwt, err
}

func getToken(ctx context.Context, apiClient client.Interface, clientID, clientSecret string) (*oauth2.Token, error) {
	req := apiClient.Post().APIPath(basicAuthEndpoint)
	endpoint := req.URL()

	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     endpoint.String(),
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	tokenContext := context.WithValue(ctx, oauth2.HTTPClient, apiClient.HTTPClient())
	return config.Token(tokenContext)
}
