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

package transport

import (
	"net/http"
)

func roundTripperWrappersForConfig(config *Config, roundTripper http.RoundTripper) http.RoundTripper {
	if len(config.UserAgent) > 0 {
		roundTripper = NewUserAgentRoundTripper(config.UserAgent, roundTripper)
	}

	// keep this wrapping as the latest possible to allow the reusage of the roundTripper during auth flow
	if config.AuthorizeWrapper != nil {
		roundTripper = config.AuthorizeWrapper(roundTripper)
	}
	return roundTripper
}

type userAgentRoundTripper struct {
	userAgent string
	next      http.RoundTripper
}

func NewUserAgentRoundTripper(userAgent string, next http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{
		userAgent: userAgent,
		next:      next,
	}
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("User-Agent")) != 0 {
		return rt.next.RoundTrip(req)
	}
	clonedReq := cloneRequest(req)
	clonedReq.Header.Set("User-Agent", rt.userAgent)
	return rt.next.RoundTrip(clonedReq)
}

// cloneRequest return a cloned version of request, used to abide to the RoundTripper contract
func cloneRequest(request *http.Request) *http.Request {
	// shallow copy of the struct
	clone := new(http.Request)
	*clone = *request

	// deep copy of the Header
	clone.Header = request.Header.Clone()
	if clone.Header == nil {
		clone.Header = make(http.Header)
	}

	return clone
}
