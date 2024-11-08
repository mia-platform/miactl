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

package transport

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportForConfig(t *testing.T) {
	wd, _ := os.Getwd()
	caFilePath, keyFilePath, _ := generateCertificates(t)
	testCases := map[string]struct {
		Config       *Config
		Err          bool
		TLS          bool
		Default      bool
		Insecure     bool
		DefaultRoots bool
	}{
		"default transport": {
			Default: true,
			Config:  &Config{},
		},
		"insecure": {
			TLS:          true,
			Insecure:     true,
			DefaultRoots: true,
			Config: &Config{
				TLSConfig: TLSConfig{
					Insecure: true,
				},
			},
		},
		"bad ca file transport": {
			Err: true,
			Config: &Config{
				TLSConfig: TLSConfig{
					CAFile: "invalid file",
				},
			},
		},
		"ca file transport": {
			TLS: true,
			Config: &Config{
				TLSConfig: TLSConfig{
					CAFile: caFilePath,
				},
			},
		},
		"wrong ca file transport": {
			Err: true,
			Config: &Config{
				TLSConfig: TLSConfig{
					CAFile: keyFilePath,
				},
			},
		},
		"not a pem file transport": {
			Err: true,
			Config: &Config{
				TLSConfig: TLSConfig{
					CAFile: path.Join(wd, "testdata", "not-a-cert.crt"),
				},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			roundTripper, err := New(testCase.Config)
			if testCase.Err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			switch {
			case testCase.Default:
				assert.Same(t, http.DefaultTransport, roundTripper)
			case !testCase.Default:
				assert.NotSame(t, http.DefaultTransport, roundTripper)
			}

			// cast roundTripper to Transport for checking properties
			transport := roundTripper.(*http.Transport)
			defaultTLSConfig := http.DefaultTransport.(*http.Transport).TLSClientConfig
			if testCase.TLS {
				assert.NotSame(t, transport.TLSClientConfig, defaultTLSConfig)
			} else {
				assert.Same(t, transport.TLSClientConfig, defaultTLSConfig)
				return
			}

			switch {
			case testCase.DefaultRoots:
				assert.Nil(t, transport.TLSClientConfig.RootCAs)
			case !testCase.DefaultRoots:
				assert.NotNil(t, transport.TLSClientConfig.RootCAs)
			}

			assert.Equal(t, testCase.Insecure, transport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}

// generateCertificates generates certificate chain for testing purposes
func generateCertificates(t *testing.T) (caFilePath, keyFilePath, certFilePath string) {
	t.Helper()
	tempDir := t.TempDir()
	notBefore := time.Now().Add(time.Minute * -5)
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Logf("failed to generate serial number: %s", err)
		t.FailNow()
	}

	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Logf("failed scdsa.GenerateKey: %s", err)
		t.FailNow()
	}

	rootTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		Subject:               pkix.Name{Organization: []string{"nil1"}},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Logf("failed createCertificate for CA: %s", err)
		t.FailNow()
	}

	caFilePath = path.Join(tempDir, "tls.ca")
	caFile, err := os.Create(caFilePath)
	if err != nil {
		t.Logf("fail to open file at %s: %s", caFilePath, err)
		t.FailNow()
	}
	defer caFile.Close()
	err = pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		t.Logf("fail to encode ca: %s", err)
		t.FailNow()
	}

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Logf("failed createLeafKey for certificate: %s", err)
		t.FailNow()
	}

	keyBytes, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		t.Logf("unable to marshal ECDSA private key: %s", err)
		t.FailNow()
	}

	keyFilePath = path.Join(tempDir, "tls.key")
	keyFile, err := os.Create(keyFilePath)
	if err != nil {
		t.Logf("fail to open file at %s: %s", keyFilePath, err)
		t.FailNow()
	}
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		t.Logf("fail to encode key: %s", err)
		t.FailNow()
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Logf("failed to generate serial number: %s", err)
		t.FailNow()
	}

	leafTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		Subject:               pkix.Name{CommonName: "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &leafTemplate, &rootTemplate, &leafKey.PublicKey, rootKey)
	if err != nil {
		t.Logf("failed createLeaf certificate: %s", err)
		t.FailNow()
	}

	certFilePath = path.Join(tempDir, "tls.crt")
	certFile, err := os.Create(certFilePath)
	if err != nil {
		t.Logf("fail to open file at %s: %s", certFilePath, err)
		t.FailNow()
	}
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		t.Logf("fail to encode certificate: %s", err)
		t.FailNow()
	}

	return caFilePath, keyFilePath, certFilePath
}
