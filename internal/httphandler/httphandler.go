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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/login"
)

type SessionHandler struct {
	url     string
	method  string
	context string
	body    io.ReadCloser
	client  *http.Client
	auth    IAuth
}

const (
	unauthorized = "401 Unauthorized"
	oktaProvider = "okta"
)

type Authenticate func() (string, error)

func NewSessionHandler(url string) (*SessionHandler, error) {
	sh := &SessionHandler{
		url: url,
	}
	return sh, nil
}

func (s *SessionHandler) WithAuthentication(url, providerID string, b login.BrowserI) *SessionHandler {
	s.auth = &Auth{
		browser:    b,
		providerID: providerID,
		url:        url,
	}
	return s
}

func (s *SessionHandler) WithBody(body io.ReadCloser) *SessionHandler {
	s.body = body
	return s
}

func (s *SessionHandler) Get() *SessionHandler {
	s.method = "GET"
	return s
}

func (s *SessionHandler) Post(body io.ReadCloser) *SessionHandler {
	s.method = "POST"
	s.WithBody(body)
	return s
}

func (s *SessionHandler) WithClient(c *http.Client) *SessionHandler {
	s.client = c
	return s
}

func (s *SessionHandler) WithContext(ctx string) *SessionHandler {
	s.context = ctx
	return s
}

func (s *SessionHandler) GetContext() string {
	return s.context
}

func HTTPClientBuilder(opts *clioptions.CLIOptions) (*http.Client, error) {
	client := &http.Client{}
	// TODO: extract CA certificate from viper config file
	if opts.CACert != "" || opts.SkipCertificate {
		transport, err := configureTransport(opts)
		if err != nil {
			return nil, fmt.Errorf("error creating custom transport: %w", err)
		}
		client.Transport = transport
	}
	return client, nil
}

func (s *SessionHandler) ExecuteRequest() (*http.Response, error) {
	httpReq, err := http.NewRequest(s.method, s.url, s.body)
	if err != nil {
		return nil, fmt.Errorf("error building the http request: %w", err)
	}
	token, err := s.auth.authenticate()
	if err != nil {
		return nil, fmt.Errorf("error retrieving token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending the http request: %w", err)
	}
	if resp.Status == unauthorized {
		newToken, err := s.auth.authenticate()
		if err != nil {
			return resp, fmt.Errorf("error refreshing token: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+newToken)
		resp, err = s.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("error resending the http request: %w", err)
		}
	}
	return resp, nil
}

func configureTransport(opts *clioptions.CLIOptions) (*http.Transport, error) {
	transport := &http.Transport{}
	tlsConfig := &tls.Config{
		MinVersion: 0x0303,
	}
	if opts.CACert != "" {
		// load the contents of the CA certificate file
		caCert, err := os.ReadFile(opts.CACert)
		if err != nil {
			return nil, fmt.Errorf("error reading CA certificate from path %s: %w", opts.CACert, err)
		}

		// create a new CertPool object and parse the CA certificate
		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		// create a new TLS configuration object and set the root CAs
		tlsConfig.RootCAs = rootCAs
	}
	if opts.SkipCertificate {
		tlsConfig.InsecureSkipVerify = true
	}

	// set the TLS configuration to the HTTP transport
	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

func ParseResponseBody(contextName string, body io.Reader, out interface{}) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	err = json.Unmarshal(bodyBytes, &out)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error unmarshaling json response: %w", err)
	}
	return nil
}

// ConfigureDefaultSessionHandler returns a session handler with default settings
func ConfigureDefaultSessionHandler(opts *clioptions.CLIOptions, contextName, uri string) (*SessionHandler, error) {
	baseURL, err := context.GetContextBaseURL(contextName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving base URL for context %s: %w", contextName, err)
	}
	// build full path URL
	fullPathURL, err := url.JoinPath(baseURL, uri)
	if err != nil {
		return nil, fmt.Errorf("error building url: %w", err)
	}
	// create a session handler object with the full path URL
	session, err := NewSessionHandler(fullPathURL)
	if err != nil {
		return nil, fmt.Errorf("error creating session handler: %w", err)
	}
	// create a new HTTP client and attach it to the session handler
	httpClient, err := HTTPClientBuilder(opts)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client: %w", err)
	}
	session.WithContext(contextName).WithClient(httpClient).WithAuthentication(baseURL, oktaProvider, login.NewDefaultBrowser())
	return session, nil
}
