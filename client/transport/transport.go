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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
)

// New return a new http.RoundTripper derived by the configuration
func New(config *Config) (http.RoundTripper, error) {
	transport := http.DefaultTransport.(*http.Transport)

	// if transport needs special configuration for TLS config create a custom one
	if config.HasCA() || config.Insecure {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			// disable gosec because will trigger G402 but we want to be able to configure this for debug purprose
			InsecureSkipVerify: config.TLSConfig.Insecure, //nolint:gosec
		}

		if config.CAFile != "" {
			certData, err := dataFromFile(config.TLSConfig.CAFile)
			if err != nil {
				return nil, err
			}
			certPool, err := certPool(certData)
			if err != nil {
				return nil, err
			}
			tlsConfig.RootCAs = certPool
		}

		// read the default transport and use its default, and then set the TLSClientConfig
		transport = &http.Transport{
			Proxy:                 transport.Proxy,
			DialContext:           transport.DialContext,
			ForceAttemptHTTP2:     transport.ForceAttemptHTTP2,
			MaxIdleConns:          transport.MaxIdleConns,
			IdleConnTimeout:       transport.IdleConnTimeout,
			TLSHandshakeTimeout:   transport.TLSHandshakeTimeout,
			ExpectContinueTimeout: transport.ExpectContinueTimeout,
			TLSClientConfig:       tlsConfig,
			DisableCompression:    config.DisableCompression,
		}
	}

	return WrappedRoundTripperForConfig(config, transport), nil
}

// dataFromFile return the read data from filePath or an error if occurred
func dataFromFile(filePath string) ([]byte, error) {
	if len(filePath) > 0 {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return []byte{}, err
		}
		return fileData, nil
	}
	return nil, nil
}

// certPool create a new cert pool starting from caData
func certPool(data []byte) (*x509.CertPool, error) {
	// if the data is empty return nil, this will allow the usage of the system trust store by default
	if len(data) == 0 {
		return nil, nil
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(data); !ok {
		return nil, errorParsingCAData(data)
	}
	return certPool, nil
}

// errorParsingCAData will return the effective error that AppendCertsFromPEM has found,
// the information is hidden by the method but it will be useful to present to the user for
// debugging resons
func errorParsingCAData(pemCerts []byte) error {
	// these checks are the ones executed inside AppendCertsFromPEM
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			return fmt.Errorf("unable to parse file as PEM")
		}

		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		certBytes := block.Bytes
		_, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
	}

	return fmt.Errorf("no valid certificate authority data")
}
