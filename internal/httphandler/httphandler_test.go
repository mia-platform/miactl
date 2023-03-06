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
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/require"
)

const testURL = "https://testurl.io/testget"

var (
	testToken   string
	testDirPath string
	client      = &http.Client{}
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
	opts := &clioptions.CLIOptions{
		APIBaseURL: testURL,
	}
	expectedReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockValidToken,
	}
	actualReq, err := RequestBuilder(opts, mockValidToken)
	require.NoError(t, err)
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
	resp, err := validReq.Get().Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request with expired token
	testToken = ""
	expReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockExpiredToken,
	}
	resp, err = expReq.Get().Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test auth error
	testToken = ""
	failAuthReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockFailAuth,
	}
	resp, err = failAuthReq.Get().Execute()
	require.Nil(t, resp)
	require.Equal(t, "error retrieving token: authentication failed", err.Error())

	// Test token refresh error
	testToken = ""
	failRefreshReq := &Request{
		url:    testURL,
		client: client,
		authFn: mockFailRefresh,
	}
	resp, err = failRefreshReq.Get().Execute()
	require.Equal(t, "401", resp.Status)
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

	opts := &clioptions.CLIOptions{
		APIBaseURL: server.URL,
		CACert:     certPath,
	}
	req, err := RequestBuilder(opts, mockValidToken)
	require.NoError(t, err)
	require.NotNil(t, req)

	resp, err := req.Get().Execute()
	require.NoError(t, err)
	require.Equal(t, "200 OK", resp.Status)
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

	testDirPath := t.TempDir()
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
		fmt.Fprintln(w, "Hello, world!")
		auth := r.Header.Get("Authorization")
		switch auth {
		case "Bearer valid_token":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	return server
}
