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
	"github.com/mia-platform/miactl/internal/cmd/login"
)

type MiaClient struct {
	request Request
	auth    IAuth
}

func NewMiaClientBuilder() *MiaClient {
	return &MiaClient{}
}

func (m *MiaClient) WithRequest(r Request) *MiaClient {
	m.request = r
	return m
}

func (m *MiaClient) WithAuthentication(providerID, url string, b login.BrowserI) *MiaClient {
	m.auth = &Auth{
		url:        url,
		browser:    b,
		providerID: providerID,
	}
	m.request.authFn = m.auth.authenticate
	return m
}
