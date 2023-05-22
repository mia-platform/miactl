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
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
)

func init() {
	if err := client.RegisterAuthProvider(NewAuthenticator); err != nil {
		panic(fmt.Sprintf("failed to register: %s", err))
	}
}

func NewAuthenticator(config *client.Config, cacheReadWriter client.AuthCacheReadWriter) client.AuthProvider {
	clonedConfig := *config

	return &authenticator{
		config:          &clonedConfig,
		cacheReadWriter: cacheReadWriter,
	}
}

type authenticator struct {
	config          *client.Config
	cacheReadWriter client.AuthCacheReadWriter
}

func (a *authenticator) Wrap(rt http.RoundTripper) http.RoundTripper {
	a.config.Transport = rt
	client, err := client.APIClientForConfig(a.config)
	if err != nil {
		fmt.Println(err)
	}

	userAuth := &userAuthenticator{
		client:   client,
		next:     rt,
		userAuth: a.cacheReadWriter,
		serverReadyHandler: func(url string) error {
			if err := open(url); err != nil {
				return fmt.Errorf("could not open the browser: %w", err)
			}
			return nil
		},
	}

	return userAuth
}
