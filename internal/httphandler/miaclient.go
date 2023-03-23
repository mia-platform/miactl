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
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
)

type MiaClient struct {
	sessionHandler SessionHandler
}

func NewMiaClientBuilder() *MiaClient {
	return &MiaClient{}
}

func (m *MiaClient) WithSessionHandler(s SessionHandler) *MiaClient {
	m.sessionHandler = s
	return m
}

func (m *MiaClient) GetSession() *SessionHandler {
	return &m.sessionHandler
}

func ConfigureDefaultMiaClient(opts *clioptions.CLIOptions, uri string) (*MiaClient, error) {

	mc := NewMiaClientBuilder()

	currentContext, err := context.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("error retrieving current context: %w", err)
	}

	session, err := ConfigureDefaultSessionHandler(opts, currentContext, uri)
	if err != nil {
		return nil, fmt.Errorf("error building default session handler: %w", err)
	}
	// attach session handler to mia client
	return mc.WithSessionHandler(*session), nil

}

func FakeMiaClient(url string) *MiaClient {
	return &MiaClient{
		sessionHandler: *FakeSessionHandler(url),
	}
}
