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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTokenError(t *testing.T) {
	config := Config{}
	_, err := config.GetToken(context.TODO())
	assert.Error(t, err, "error expected from config without appIOd")

	config.AppID = "appid"
	_, err = config.GetToken(context.TODO())
	assert.Error(t, err, "error expected from config without client")
}

func TestRefreshTokenError(t *testing.T) {
	config := Config{}
	_, err := config.RefreshToken(context.TODO(), "")
	assert.Error(t, err, "error expected from config without refreshToken")

	_, err = config.RefreshToken(context.TODO(), "refresh-token")
	assert.Error(t, err, "error expected from config without client")
}

func TestNewListener(t *testing.T) {
	listener, err := newListener(nil)
	require.NoError(t, err)
	require.NotNil(t, listener)

	errListener, err := newListener([]string{listener.Addr().String()})
	assert.Error(t, err)
	assert.Nil(t, errListener)

	newListener, err := newListener([]string{listener.Addr().String(), "127.0.0.1:0"})
	require.NoError(t, err)
	require.NotNil(t, newListener)
	listener.Close()
	newListener.Close()
}

func TestLocalServer(t *testing.T) {
	channel := make(chan *authResponse)
	handler := &httpHandler{
		startFlowURL:    "https://example.com",
		responseChannel: channel,
	}

	testCases := map[string]struct {
		recorder           *httptest.ResponseRecorder
		request            *http.Request
		expectedStatusCode int
		async              bool
	}{
		"login error": {
			recorder: httptest.NewRecorder(),
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, callbackEndpointString, nil)
				query := request.URL.Query()
				query.Set("error", "login error")
				request.URL.RawQuery = query.Encode()
				return request
			}(),
			expectedStatusCode: http.StatusInternalServerError,
			async:              true,
		},
		"unknown request": {
			recorder:           httptest.NewRecorder(),
			request:            httptest.NewRequest(http.MethodGet, "/foo", nil),
			expectedStatusCode: http.StatusNotFound,
			async:              true,
		},
		"redirect on root": {
			recorder:           httptest.NewRecorder(),
			request:            httptest.NewRequest(http.MethodGet, "/", nil),
			expectedStatusCode: http.StatusFound,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			if testCase.async {
				go func() {
					handler.ServeHTTP(testCase.recorder, testCase.request)
				}()
				response := <-channel
				assert.Error(t, response.Err)
			} else {
				handler.ServeHTTP(testCase.recorder, testCase.request)
			}
			httpResponse := testCase.recorder.Result()
			httpResponse.Body.Close()
			assert.Equal(t, testCase.expectedStatusCode, httpResponse.StatusCode)
		})
	}
}
