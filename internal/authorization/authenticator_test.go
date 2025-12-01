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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mia-platform/miactl/internal/client"
)

func TestNewAuthenticator(t *testing.T) {
	config := &client.Config{}
	authProvider := NewAuthenticator(config, &testAuthCacheProvider{}, client.AuthConfig{})
	assert.NotNil(t, authProvider)
	assert.IsType(t, &authenticator{}, authProvider)
}

func TestAuthenticatorWrapping(t *testing.T) {
	authProvider := &authenticator{
		config:          &client.Config{Host: "http://example.com"},
		cacheReadWriter: &testAuthCacheProvider{},
	}

	wrappedRt := authProvider.Wrap(http.DefaultTransport)
	assert.NotNil(t, wrappedRt)
	assert.IsType(t, &userAuthenticator{}, wrappedRt)
}
