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
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	testURL, err := url.Parse("http://host")
	if err != nil {
		require.NoError(t, err)
	}
	testContentType := "text/html"
	testCases := map[string]struct {
		client               *APIClient
		expectedAcceptHeader string
	}{
		"with accept content type": {
			client:               newAPIClient(testURL, contentConfig{AcceptContentTypes: testContentType}, http.DefaultClient),
			expectedAcceptHeader: testContentType,
		},
		"without contetn type": {
			client:               newAPIClient(testURL, contentConfig{ContentType: testContentType}, http.DefaultClient),
			expectedAcceptHeader: fmt.Sprintf("%s, */*", testContentType),
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			APIClient := NewRequest(testCase.client)
			actualAcceptHeader := APIClient.headers.Get("accept")
			assert.Equal(t, testCase.expectedAcceptHeader, actualAcceptHeader)
		})
	}
}

func TestSetParams(t *testing.T) {
	r := (&Request{}).SetParam("foo", "bar")
	assert.Equal(t, r.params, url.Values{"foo": []string{"bar"}})

	r.SetParam("baz", "1", "2")
	assert.Equal(t, r.params, url.Values{"foo": []string{"bar"}, "baz": []string{"1", "2"}})
}

func TestSetAPIPath(t *testing.T) {
	r := (&Request{})

	validPath := "/valid/path"
	r.APIPath(validPath)
	assert.NoError(t, r.Error())
	assert.Equal(t, r.apiPath, fmt.Sprintf("%s/", "/valid/path"))
	r.APIPath(":invalid-url")
	assert.Error(t, r.Error())

	// once an error is register no other changes can be made
	r.APIPath(validPath)
	assert.Error(t, r.Error())
}

func TestPreflightCheck(t *testing.T) {
	testCases := map[string]struct {
		request *Request
		err     bool
	}{
		"correct GET": {
			request: (&Request{}).SetVerb("GET"),
		},
		"correct POST": {
			request: (&Request{}).SetVerb("POST").Body([]byte("hello")),
		},
		"empty verb": {
			request: &Request{},
			err:     true,
		},
		"get with body": {
			request: (&Request{}).SetVerb("GET").Body([]byte("hello")),
			err:     true,
		},
		"empty body": {
			request: (&Request{}).SetVerb("POST").Body([]byte{}),
			err:     true,
		},
		"valid verb and body but preexisting error": {
			request: (&Request{err: fmt.Errorf("")}).SetVerb("GET"),
			err:     true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := testCase.request.preflightCheck()
			switch testCase.err {
			case true:
				assert.Error(t, err)
			case false:
				assert.NoError(t, err)
			}
		})
	}
}
