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

package client

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type AuthCacheReadWriter interface {
	ReadJWTToken() *oauth2.Token
	WriteJWTToken(*oauth2.Token)
}

type AuthProvider interface {
	// Wrap allow the AuthProvider to add authorization functionality on
	// a modified RoundTripper and to add the appropriate Authorization header to the request
	Wrap(http.RoundTripper) http.RoundTripper
}

// AuthProviderCreator is a function that return an AuthProvider
type AuthProviderCreator func(*Config, AuthCacheReadWriter) AuthProvider

var authProvidersLock sync.Mutex
var authProvider AuthProviderCreator

func RegisterAuthProvider(ap AuthProviderCreator) error {
	authProvidersLock.Lock()
	defer authProvidersLock.Unlock()

	if authProvider != nil {
		return fmt.Errorf("another auth provider is already registred")
	}

	authProvider = ap
	return nil
}
