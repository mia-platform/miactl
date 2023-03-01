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

package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

type AssertionFn func(t *testing.T, req *http.Request)

type Response struct {
	Assertions AssertionFn
	Body       string
	Status     int
}
type Responses []Response

func CreateTestClient(t *testing.T, url string) *jsonclient.Client {
	t.Helper()
	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
		Headers: jsonclient.Headers{
			"cookie": "sid=my-random-sid",
		},
	})
	require.NoError(t, err, "error creating client")
	return client
}

func CreateTestResponseServer(t *testing.T, assertions AssertionFn, responseBody string, statusCode int) *httptest.Server {
	t.Helper()
	responses := []Response{
		{Assertions: assertions, Body: responseBody, Status: statusCode},
	}
	return CreateMultiTestResponseServer(t, responses)
}

func CreateMultiTestResponseServer(t *testing.T, responses Responses) *httptest.Server {
	t.Helper()
	var usage int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if usage >= len(responses) {
			t.Fatalf("Unexpected HTTP request, provided %d handler, received call #%d.", len(responses), usage+1)
		}

		response := responses[usage]
		usage++
		if response.Assertions != nil {
			response.Assertions(t, req)
		}

		w.WriteHeader(response.Status)
		var responseBytes []byte
		if response.Body != "" {
			responseBytes = []byte(response.Body)
		}
		w.Write(responseBytes)
	}))
}

func ReadTestData(t *testing.T, fileName string) string {
	t.Helper()

	fileContent, err := os.ReadFile(fmt.Sprintf("../testdata/%s", fileName))
	require.NoError(t, err)
	return string(fileContent)
}
