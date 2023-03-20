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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/stretchr/testify/require"
)

const (
	testBaseURL  = "test.url"
	testProvider = "testProvider"
	testURI      = "/test"
)

var (
	testToken     string
	testDirPath   string
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

func TestNewSessionHandler(t *testing.T) {
	expected := &SessionHandler{
		url: testURI,
	}
	actualReq, err := NewSessionHandler(testURI)
	require.NoError(t, err)
	require.Equal(t, expected.url, actualReq.url)
}

func TestHttpClientBuilder(t *testing.T) {
	certPath, _, err := generateMockCert(t)
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
	server := createMockServer()
	server.Start()
	defer server.Close()

	// Test request with valid token
	testToken = ""
	validAuth := mockValidToken{}
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
	expAuth := mockExpiredToken{}
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
	failAuth := mockFailAuth{}
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
	failRefresh := mockFailRefresh{}
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
	certPath, keyPath, err := generateMockCert(t)
	if err != nil {
		t.Fatalf("unexpected error")
	}

	// create mock server
	server := createMockServer()

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

	a := mockValidToken{}
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

func generateMockCert(t *testing.T) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")}, // IP SAN for 127.0.0.1
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	testDirPath = t.TempDir()
	testCertPath := path.Join(testDirPath, "testcert.pem")
	certOut, err := os.Create(testCertPath)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	testKeyPath := path.Join(testDirPath, "testkey.pem")
	keyOut, err := os.Create(testKeyPath)
	if err != nil {
		panic(err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return testCertPath, testKeyPath, nil
}

func createMockServer() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		switch auth {
		case "Bearer valid_token":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
		w.Write([]byte{})
	}))
	return server
}
