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
	"github.com/stretchr/testify/require"
)

const testURL = "https://testurl.io/testget"

var (
	testToken     string
	testDirPath   string
	defaultClient = &http.Client{}
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

func TestWithClient(t *testing.T) {
	req := &Request{}
	req.WithClient(defaultClient)
	require.NotNil(t, req.client)
	require.Equal(t, req.client, defaultClient)
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
	opts := &clioptions.CLIOptions{
		APIBaseURL: testURL,
	}
	a := mockValidToken{}
	expectedReq := &Request{
		url:    testURL,
		authFn: a.authenticate,
	}
	actualReq, err := RequestBuilder(opts, a.authenticate)
	require.NoError(t, err)
	require.Equal(t, expectedReq.url, actualReq.url)
	require.NotNil(t, actualReq.authFn)
}

func TestHttpClientBuilder(t *testing.T) {
	certPath, _, err := generateMockCert(t)
	require.NoError(t, err)

	// Test default client
	opts := &clioptions.CLIOptions{}
	client, err := httpClientBuilder(opts)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, defaultClient, client)

	// Test client with cert
	opts2 := &clioptions.CLIOptions{
		CACert: certPath,
	}
	require.NoError(t, err)
	client, err = httpClientBuilder(opts2)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotEqual(t, http.DefaultTransport, client.Transport)

	// Test client with skip cert validation
	opts3 := &clioptions.CLIOptions{
		SkipCertificate: true,
	}
	require.NoError(t, err)
	client, err = httpClientBuilder(opts3)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotEqual(t, http.DefaultTransport, client.Transport)

}

func TestExecute(t *testing.T) {
	server := createMockServer()
	server.Start()
	defer server.Close()

	// Test request with valid token
	testToken = ""
	validAuth := mockValidToken{}
	validReq := &Request{
		url:    server.URL,
		client: defaultClient,
		authFn: validAuth.authenticate,
	}

	mc := NewMiaClientBuilder().WithRequest(*validReq)

	resp, err := mc.request.Get().Execute()
	require.Nil(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// Test request with expired token
	testToken = ""
	expAuth := mockExpiredToken{}
	expReq := &Request{
		url:    server.URL,
		client: defaultClient,
		authFn: expAuth.authenticate,
	}
	expReq.WithClient(defaultClient)
	resp, err = expReq.Get().Execute()
	require.Nil(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// Test auth error
	testToken = ""
	failAuth := mockFailAuth{}
	failAuthReq := &Request{
		url:    server.URL,
		client: defaultClient,
		authFn: failAuth.authenticate,
	}
	failAuthReq.WithClient(defaultClient)
	resp, err = failAuthReq.Get().Execute()
	require.Nil(t, resp)
	require.Equal(t, "error retrieving token: authentication failed", err.Error())

	// Test token refresh error
	testToken = ""
	failRefresh := mockFailRefresh{}
	failRefreshReq := &Request{
		url:    server.URL,
		client: defaultClient,
		authFn: failRefresh.authenticate,
	}
	failRefreshReq.WithClient(defaultClient)
	resp, err = failRefreshReq.Get().Execute()
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

	opts := &clioptions.CLIOptions{
		APIBaseURL: server.URL,
		CACert:     certPath,
	}
	req, err := RequestBuilder(opts, a.authenticate)
	require.NoError(t, err)
	require.NotNil(t, req)

	client, err := httpClientBuilder(opts)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err := req.WithClient(client).Get().Execute()
	require.NoError(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// test skip certificate validation
	opts2 := &clioptions.CLIOptions{
		APIBaseURL:      server.URL,
		SkipCertificate: true,
	}
	req, err = RequestBuilder(opts2, a.authenticate)
	require.NoError(t, err)
	require.NotNil(t, req)

	client, err = httpClientBuilder(opts2)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err = req.WithClient(client).Get().Execute()
	require.NoError(t, err)
	require.Equal(t, "200 OK", resp.Status)

	// test fail certificate validation
	opts3 := &clioptions.CLIOptions{
		APIBaseURL: server.URL,
	}
	req, err = RequestBuilder(opts3, a.authenticate)
	require.NoError(t, err)
	require.NotNil(t, req)

	client, err = httpClientBuilder(opts3)
	require.NoError(t, err)
	require.NotNil(t, client)

	resp, err = req.WithClient(client).Get().Execute()
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
