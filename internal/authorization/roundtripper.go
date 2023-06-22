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

	"github.com/mia-platform/miactl/internal/netutil"
	"golang.org/x/oauth2"
)

type accessTokenFunc func() (*oauth2.Token, error)

func roundTrip(req *http.Request, next http.RoundTripper, fn accessTokenFunc) (*http.Response, error) {
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
		return next.RoundTrip(req)
	}

	accessToken, err := fn()
	if err != nil {
		return nil, err
	}

	clonedReq := netutil.CloneRequest(req)
	accessToken.SetAuthHeader(clonedReq)
	reqBodyClosed = true
	return next.RoundTrip(clonedReq)
}
