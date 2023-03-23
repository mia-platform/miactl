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
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mia-platform/miactl/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	testBaseURL  = "test.url"
	testProvider = "testProvider"
	testURI      = "/test"
	testContext  = "test-context"
)

var (
	testToken     string
	defaultClient = &http.Client{}
)

func TestWithBody(t *testing.T) {
	req := &SessionHandler{}
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

func TestWithClient(t *testing.T) {
	req := &SessionHandler{}
	req.WithClient(defaultClient)
	require.NotNil(t, req.client)
	require.Equal(t, req.client, defaultClient)
}

func TestGet(t *testing.T) {
	req := &SessionHandler{}
	req.Get()
	require.Equal(t, "GET", req.method)
}

func TestPost(t *testing.T) {
	req := &SessionHandler{}
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

func TestWithAuthentication(t *testing.T) {
	session := &SessionHandler{}
	browser := &login.Browser{}
	session.WithAuthentication(testBaseURL, testProvider, browser)
	expectedSession := &SessionHandler{
		auth: &Auth{
			url:        testBaseURL,
			providerID: testProvider,
			browser:    browser,
		},
	}
	require.Equal(t, expectedSession, session)
}

func TestGetContext(t *testing.T) {
	session := &SessionHandler{
		context: testContext,
	}
	context := session.GetContext()
	require.Equal(t, context, session.context)
}

func TestNewSessionHandler(t *testing.T) {
	expected := &SessionHandler{
		url: testURI,
	}
	actualReq, err := NewSessionHandler(testURI)
	require.NoError(t, err)
	require.Equal(t, expected.url, actualReq.url)
}

func TestHttpClientBuilder(t *testing.T) {
	certPath, _, err := testutils.GenerateMockCert(t)
	require.NoError(t, err)

	// Test default client
	opts := &clioptions.CLIOptions{}
	client, err := HTTPClientBuilder(opts)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, defaultClient, client)

	// Test client with cert
	opts2 := &clioptions.CLIOptions{
		CACert: certPath,
	}
	require.NoError(t, err)
	client, err = HTTPClientBuilder(opts2)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotEqual(t, http.DefaultTransport, client.Transport)

	// Test client with skip cert validation
	opts3 := &clioptions.CLIOptions{
		SkipCertificate: true,
	}
	require.NoError(t, err)
	client, err = HTTPClientBuilder(opts3)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotEqual(t, http.DefaultTransport, client.Transport)

}

func TestExecuteRequest(t *testing.T) {
	server := testutils.CreateMockServer()
	server.Start()
	defer server.Close()

	// Test request with valid token
	testToken = ""
	validAuth := testutils.MockValidToken{}
	validSession := &SessionHandler{
		url:    server.URL,
		client: defaultClient,
		auth:   &validAuth,
	}

	mc := NewMiaClientBuilder().WithSessionHandler(*validSession)

	resp, err := mc.sessionHandler.Get().ExecuteRequest()
	require.Nil(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// Test request with expired token
	testToken = ""
	expAuth := testutils.MockExpiredToken{}
	expiredSession := &SessionHandler{
		url:    server.URL,
		client: defaultClient,
		auth:   &expAuth,
	}
	expiredSession.WithClient(defaultClient)
	resp, err = expiredSession.Get().ExecuteRequest()
	require.Nil(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// Test auth error
	testToken = ""
	failAuth := testutils.MockFailAuth{}
	failedSession := &SessionHandler{
		url:    server.URL,
		client: defaultClient,
		auth:   &failAuth,
	}
	failedSession.WithClient(defaultClient)
	resp, err = failedSession.Get().ExecuteRequest()
	require.Nil(t, resp)
	require.Equal(t, "error retrieving token: authentication failed", err.Error())

	// Test token refresh error
	testToken = ""
	failRefresh := testutils.MockFailRefresh{}
	failRefreshSession := &SessionHandler{
		url:    server.URL,
		client: defaultClient,
		auth:   &failRefresh,
	}
	failRefreshSession.WithClient(defaultClient)
	resp, err = failRefreshSession.Get().ExecuteRequest()
	require.Equal(t, unauthorized, resp.Status)
	require.Equal(t, "error refreshing token: authentication failed", err.Error())
}

func TestReqWithCustomTransport(t *testing.T) {
	// create mock certificate
	certPath, keyPath, err := testutils.GenerateMockCert(t)
	if err != nil {
		t.Fatalf("unexpected error")
	}

	// create mock server
	server := testutils.CreateMockServer()

	// load certificate and start TLS
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server.StartTLS()
	defer server.Close()

	a := testutils.MockValidToken{}
	session := &SessionHandler{
		url:    server.URL,
		client: defaultClient,
		auth:   &a,
	}

	opts := &clioptions.CLIOptions{
		CACert: certPath,
	}

	client, err := HTTPClientBuilder(opts)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err := session.WithClient(client).Get().ExecuteRequest()
	require.NoError(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// test skip certificate validation
	opts2 := &clioptions.CLIOptions{
		SkipCertificate: true,
	}

	client, err = HTTPClientBuilder(opts2)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err = session.WithClient(client).Get().ExecuteRequest()
	require.NoError(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// test fail certificate validation
	opts3 := &clioptions.CLIOptions{
		APIBaseURL: server.URL,
	}

	client, err = HTTPClientBuilder(opts3)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err = session.WithClient(client).Get().ExecuteRequest()
	require.Nil(t, resp)
	require.Error(t, err)
}

func TestParseResponseBody(t *testing.T) {
	var out testutils.Test

	// valid json body
	body := bytes.NewReader([]byte(`invalid json`))
	expectedOut := testutils.Test{}
	err := ParseResponseBody(testContext, body, &out)
	require.Equal(t, expectedOut, out)
	require.Error(t, err)

	// invalid json
	body = bytes.NewReader([]byte(`{"key": "value"}`))
	expectedOut = testutils.Test{
		Key: "value",
	}
	err = ParseResponseBody(testContext, body, &out)
	require.Equal(t, expectedOut, out)
	require.NoError(t, err)

}

func TestConfigureDefaultSessionHandler(t *testing.T) {
	opts := &clioptions.CLIOptions{}
	viper.SetConfigType("yaml")
	config := `contexts:
  test-context:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	// valid session
	expectedSession := &SessionHandler{
		url:     "http://url/test",
		context: testContext,
		client:  defaultClient,
		auth: &Auth{
			url:        "http://url",
			providerID: oktaProvider,
			browser:    login.Browser{},
		},
	}
	session, err := ConfigureDefaultSessionHandler(opts, testContext, testURI)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.EqualValues(t, expectedSession, session)

	// invalid context
	session, err = ConfigureDefaultSessionHandler(opts, "wrong-context", testURI)
	require.Nil(t, session)
	require.Error(t, err)

}
