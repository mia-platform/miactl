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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRoundTripper struct {
	Request *http.Request
	Err     error
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Request = req
	return nil, rt.Err
}

func TestUserAgentRoundTripper(t *testing.T) {
	rt := &testRoundTripper{}

	t.Run("user agent header already present", func(t *testing.T) {
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Set("User-Agent", "other")
		// turn off bodyclose, because we don't have body to close here...
		NewUserAgentRoundTripper("test", rt).RoundTrip(req) //nolint:bodyclose
		require.NotNil(t, rt.Request)

		rtRequest := rt.Request
		assert.Same(t, rtRequest, req)
		assert.Equal(t, rtRequest.Header.Get("User-Agent"), "other")
	})

	t.Run("missing user agent in request", func(t *testing.T) {
		req := &http.Request{}
		// turn off bodyclose, because we don't have body to close here...
		NewUserAgentRoundTripper("test", rt).RoundTrip(req) //nolint:bodyclose
		require.NotNil(t, rt.Request)

		rtRequest := rt.Request
		assert.NotSame(t, rtRequest, req)
		assert.Equal(t, rtRequest.Header.Get("User-Agent"), "test")
	})
}

type testAuthTripper struct{}

func (rt *testAuthTripper) RoundTrip(req *http.Request) (*http.Response, error) { return nil, nil }

func TestRoundTripperWrapping(t *testing.T) {
	baseTransport := &http.Transport{}

	testCases := map[string]struct {
		config       *Config
		expectedType interface{}
	}{
		"empty config": {
			config:       &Config{},
			expectedType: baseTransport,
		},
		"user agent": {
			config:       &Config{UserAgent: "foo"},
			expectedType: &userAgentRoundTripper{},
		},
		"auth wrapper": {
			config: &Config{
				AuthorizeWrapper: func(rt http.RoundTripper) http.RoundTripper { return &testAuthTripper{} },
			},
			expectedType: &testAuthTripper{},
		},
		"both config, return auth wrapper": {
			config: &Config{
				UserAgent:        "foo",
				AuthorizeWrapper: func(rt http.RoundTripper) http.RoundTripper { return &testAuthTripper{} },
			},
			expectedType: &testAuthTripper{},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			wrappedTransform := roundTripperWrappersForConfig(testCase.config, baseTransport)
			assert.IsType(t, testCase.expectedType, wrappedTransform)
		})
	}
}
