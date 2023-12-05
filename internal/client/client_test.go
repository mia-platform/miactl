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
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	testCases := map[string]struct {
		URLString           string
		ContentConfig       contentConfig
		HTTPCLient          *http.Client
		OverrideContentType bool
	}{
		"default creation": {
			URLString: "https://test/prefix",
			ContentConfig: contentConfig{
				ContentType: "text/html",
			},
			HTTPCLient: http.DefaultClient,
		},
		"override content type if empty": {
			URLString:           "https://test/prefix",
			ContentConfig:       contentConfig{},
			HTTPCLient:          http.DefaultClient,
			OverrideContentType: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			url, err := url.Parse(testCase.URLString)
			require.NoError(t, err)

			client := newAPIClient(url, testCase.ContentConfig, testCase.HTTPCLient)
			assert.Equal(t, testCase.HTTPCLient, client.client)
			assert.True(t, strings.HasSuffix(client.baseURL.Path, "/"))
			if testCase.OverrideContentType {
				assert.Equal(t, defaultContentType, client.contentConfig.ContentType)
			} else {
				assert.Equal(t, testCase.ContentConfig.ContentType, client.contentConfig.ContentType)
			}
		})
	}
}

func TestRequestSuccess(t *testing.T) {
	testServer := testServerEnv(t, 200)
	defer testServer.Close()

	restClient, err := APIClientForConfig(&Config{Host: testServer.URL})
	require.NoError(t, err)

	req := restClient.Get().APIPath("test")
	response, err := req.Do(context.TODO())

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, []byte("{}"), response.body)
}

func TestRequestError(t *testing.T) {
	testServer := testServerEnv(t, 200)
	// close immediately for returning a network error
	testServer.Close()

	restClient, err := APIClientForConfig(&Config{Host: testServer.URL})
	require.NoError(t, err)

	req := restClient.Get().APIPath("test")
	response, err := req.Do(context.TODO())

	require.Error(t, err)
	require.Nil(t, response)
}

func TestRequestServerError(t *testing.T) {
	testServer := testServerEnv(t, 400)
	defer testServer.Close()

	restClient, err := APIClientForConfig(&Config{Host: testServer.URL})
	require.NoError(t, err)

	req := restClient.Get().APIPath("test")
	response, err := req.Do(context.TODO())

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 400, response.StatusCode())
	assert.Error(t, response.Error())
}

func TestRequestServer5xx(t *testing.T) {
	testServer := testServerEnv(t, 500)
	defer testServer.Close()

	restClient, err := APIClientForConfig(&Config{Host: testServer.URL})
	require.NoError(t, err)

	req := restClient.Get().APIPath("test")
	response, err := req.Do(context.TODO())

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode())
	assert.Error(t, response.Error())
}

func testServerEnv(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("{}"))
	})

	testServer := httptest.NewServer(fakeHandler)
	return testServer
}
