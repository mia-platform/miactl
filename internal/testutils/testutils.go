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

package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"
)

var TestToken = ""

type Test struct {
	Key string `json:"key"`
}

type MockValidToken struct{}
type MockExpiredToken struct{}
type MockFailAuth struct{}
type MockFailRefresh struct{}

func (a *MockValidToken) Authenticate() (string, error) {
	return "valid_token", nil
}

func (a *MockExpiredToken) Authenticate() (string, error) {
	if TestToken == "" {
		TestToken = "expired_token"
	} else {
		TestToken = "valid_token"
	}
	return TestToken, nil
}

func (a *MockFailAuth) Authenticate() (string, error) {
	return "", fmt.Errorf("authentication failed")
}

func (a *MockFailRefresh) Authenticate() (string, error) {
	if TestToken == "" {
		TestToken = "expired_token"
		return TestToken, nil
	}
	return "", fmt.Errorf("authentication failed")
}

// GenerateMockCert generates a fake certificate for testing purposes
func GenerateMockCert(t *testing.T) (string, string, error) {
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
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return "", "", err
	}

	testKeyPath := path.Join(testDirPath, "testkey.pem")
	keyOut, err := os.Create(testKeyPath)
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return "", "", err
	}

	return testCertPath, testKeyPath, nil
}

// CreateMockServer creates a mock server for testing purposes
func CreateMockServer() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/notfound" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			auth := r.Header.Get("Authorization")
			switch auth {
			case "Bearer valid_token":
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
		if r.RequestURI == "/invalidbody" {
			_, err := w.Write([]byte(`invalid json`))
			if err != nil {
				panic(err)
			}
		} else if r.RequestURI == "/getprojects" {
			_, err := w.Write([]byte(`[{"_id": "123"}]`))
			if err != nil {
				panic(err)
			}
		}
	}))
	return server
}
