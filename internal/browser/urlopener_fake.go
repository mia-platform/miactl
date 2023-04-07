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
	"fmt"
	"net/http"
	"testing"
)

// FakeURLOpener fake implementation for BrowserOpener, use for testing the functionality
// and easily mock responses to the callback
type FakeURLOpener struct {
	t           *testing.T
	code        string
	state       string
	callbackURL string
}

func NewFakeURLOpener(t *testing.T, code, state, callbackURL string) *FakeURLOpener {
	return &FakeURLOpener{
		t:           t,
		code:        code,
		state:       state,
		callbackURL: callbackURL,
	}
}

// Open the function will call in a go routine the CallBackURL with Code and State of the mocked opener
func (f FakeURLOpener) Open(_ string) error {
	f.t.Helper()
	go func() {
		// no need to check things, is only for testing
		// nolint errcheck
		http.DefaultClient.Get(fmt.Sprintf("http://%s/oauth/callback?code=%s&state=%s", f.callbackURL, f.code, f.state))
	}()
	return nil
}
