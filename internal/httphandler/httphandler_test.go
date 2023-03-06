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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/mia-platform/miactl/old/sdk"
	"github.com/stretchr/testify/require"
)

const testURL = "https://testurl.io/testget"

var (
	testToken string
	client    = &http.Client{}
)

func TestWithBody(t *testing.T) {
	req := &Request{}
	values := map[string]string{"key": "value"}
	jsonValues, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	body := bytes.NewBuffer(jsonValues)
	wrappedBody := io.NopCloser(body)
	req.WithBody(wrappedBody)
	require.Equal(t, wrappedBody, req.body)
}

func TestGet(t *testing.T) {
	req := &Request{}
	req.Get()
	require.Equal(t, "GET", req.method)
}

func TestPost(t *testing.T) {
	req := &Request{}
	values := map[string]string{"key": "value"}
	jsonValues, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	body := bytes.NewBuffer(jsonValues)
	wrappedBody := io.NopCloser(body)
	req.Post(wrappedBody)
	require.Equal(t, "POST", req.method)
	require.Equal(t, wrappedBody, req.body)
}

func TestRequestBuilder(t *testing.T) {
	opts := &sdk.Options{
		APIBaseURL: testURL,
	}
	expectedReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockValidToken,
	}
	actualReq := RequestBuilder(*opts, mockValidToken)
	require.Equal(t, expectedReq.url, actualReq.url)
	require.NotNil(t, actualReq.authFn)
}

func TestExecute(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", testURL,
		func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error
			if req.Header.Get("Authorization") != "Bearer valid_token" {
				resp, err = httpmock.NewJsonResponse(401, map[string]interface{}{
					"authorized": "false",
				})
			} else {
				resp, err = httpmock.NewJsonResponse(200, map[string]interface{}{
					"authorized": "true",
				})
			}
			return resp, err
		},
	)

	// Test request with valid token
	testToken = ""
	validReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockValidToken,
	}
	resp, err := validReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request with expired token
	testToken = ""
	expReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockExpiredToken,
	}
	resp, err = expReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test auth error
	testToken = ""
	failAuthReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockFailAuth,
	}
	resp, err = failAuthReq.Execute()
	require.Nil(t, resp)
	require.Equal(t, "error retrieving token: authentication failed", err.Error())

	// Test token refresh error
	testToken = ""
	failRefreshReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockFailRefresh,
	}
	resp, err = failRefreshReq.Execute()
	require.Equal(t, "401", resp.Status)
	require.Equal(t, "error refreshing token: authentication failed", err.Error())
}

func mockValidToken(url string) (string, error) {
	return "valid_token", nil
}

func mockExpiredToken(url string) (string, error) {
	if testToken == "" {
		testToken = "expired_token"
	} else {
		testToken = "valid_token"
	}
	return testToken, nil
}

func mockFailAuth(url string) (string, error) {
	return "", fmt.Errorf("authentication failed")
}

func mockFailRefresh(url string) (string, error) {
	if testToken == "" {
		testToken = "expired_token"
		return testToken, nil
	}
	return "", fmt.Errorf("authentication failed")
}
