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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testAuthProviderCreator(*Config, AuthCacheReadWriter, AuthConfig) AuthProvider {
	return &testAuthProvider{}
}

type testAuthProvider struct{}

func (ap *testAuthProvider) Wrap(rt http.RoundTripper) http.RoundTripper { return rt }

func TestSetAuthorizationProvider(t *testing.T) {
	authProvider = nil
	assert.NoError(t, RegisterAuthProvider(testAuthProviderCreator))
	assert.Error(t, RegisterAuthProvider(testAuthProviderCreator))
	authProvider = nil
}
