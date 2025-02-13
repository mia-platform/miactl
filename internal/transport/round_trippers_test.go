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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/mia-platform/miactl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRoundTripper struct {
	Request *http.Request
	Err     error
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.Request = req
	return &http.Response{
		Status:     "OK",
		StatusCode: http.StatusOK,
		Header: map[string][]string{
			"X-Test-Header": {"test value"},
		},
	}, rt.Err
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

func (rt *testAuthTripper) RoundTrip(_ *http.Request) (*http.Response, error) { return nil, nil }

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
		"verbose": {
			config:       &Config{Verbose: true},
			expectedType: &debugRoundTripper{},
		},
		"user agent": {
			config:       &Config{UserAgent: "foo"},
			expectedType: &userAgentRoundTripper{},
		},
		"auth wrapper": {
			config: &Config{
				AuthorizeWrapper: func(_ http.RoundTripper) http.RoundTripper { return &testAuthTripper{} },
			},
			expectedType: &testAuthTripper{},
		},
		"both config, return auth wrapper": {
			config: &Config{
				UserAgent:        "foo",
				AuthorizeWrapper: func(_ http.RoundTripper) http.RoundTripper { return &testAuthTripper{} },
			},
			expectedType: &testAuthTripper{},
		},
		"all config, return auth wrapper": {
			config: &Config{
				UserAgent:        "foo",
				Verbose:          true,
				AuthorizeWrapper: func(_ http.RoundTripper) http.RoundTripper { return &testAuthTripper{} },
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

func TestRedactSensibleHeaders(t *testing.T) {
	testCases := []struct {
		key      string
		value    string
		expected string
	}{
		{
			key:      "NoSensible",
			value:    "test",
			expected: "test",
		},
		{
			key:      "Authorization",
			value:    "Basic test",
			expected: "Basic REDACTED",
		},
		{
			key:      "authorization",
			value:    "Bearer token",
			expected: "Bearer REDACTED",
		},
		{
			key:      "authorization",
			value:    "test",
			expected: "REDACTED",
		},
		{
			key:      "authorization",
			value:    "digest",
			expected: "digest",
		},
		{
			key:      "authorization",
			value:    "",
			expected: "",
		},
	}

	for _, testCase := range testCases {
		maskedValue := maskSensibleHeaderValue(testCase.key, testCase.value)
		assert.Equal(t, testCase.expected, maskedValue)
	}
}

func TestDebugRoundTripper(t *testing.T) {
	t.Parallel()

	testURL := "https://127.0.0.1:8080/a/request/url/"
	request := httptest.NewRequest(http.MethodGet, testURL, nil)
	request.Header = map[string][]string{
		"Authorization": {"Bearer token"},
		"X-Test-Header": {"test value"},
	}

	// test without logger must not break anything
	rt := &testRoundTripper{}
	NewDebugRoundTripper(rt).RoundTrip(request) //nolint: bodyclose

	testCases := map[string]struct {
		logLevel            int
		expectedOutputLines []string
	}{
		"Logger with level 5": {
			logLevel:            5,
			expectedOutputLines: []string{},
		},
		"Logger with level 6": {
			logLevel: 6,
			expectedOutputLines: []string{
				fmt.Sprintf("%s, %s", request.Method, request.URL.String()),
				"Response Status: OK in 0 milliseconds",
			},
		},
		"Logger with level 7": {
			logLevel: 7,
			expectedOutputLines: []string{
				fmt.Sprintf("%s, %s", request.Method, request.URL.String()),
				"Response Status: OK in 0 milliseconds",
				"Response Headers:",
				"Authorization: Bearer REDACTED",
				"Response Headers:",
			},
		},
		"Logger with level 10": {
			logLevel: 10,
			expectedOutputLines: []string{
				fmt.Sprintf("%s, %s", request.Method, request.URL.String()),
				"Response Status: OK in 0 milliseconds",
				"Response Headers:",
				"Authorization: Bearer REDACTED",
				"Response Headers:",
				fmt.Sprintf("curl -v -X%s", request.Method),
				fmt.Sprintf("'%s'", request.URL.String()),
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// execute the round tripper
			buffer := bytes.NewBuffer(nil)
			logger := logger.NewTestLogger(buffer, testCase.logLevel)
			rt := &testRoundTripper{}
			contextRequest := request.Clone(logr.NewContext(t.Context(), logger))
			NewDebugRoundTripper(rt).RoundTrip(contextRequest) //nolint: bodyclose
			loggedOutput := buffer.String()

			if len(testCase.expectedOutputLines) == 0 {
				require.Equal(t, "", loggedOutput)
			}

			for _, expectedString := range testCase.expectedOutputLines {
				assert.True(t, strings.Contains(loggedOutput, expectedString), "%s not found", expectedString)
			}
		})
	}
}
