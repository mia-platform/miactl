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
	"github.com/mia-platform/miactl/internal/testutils"
)

// SessionHandler is the type spec for miactl HTTP sessions
type SessionHandler struct {
	url     string
	method  string
	context string
	body    io.Reader
	client  *http.Client
	auth    IAuth
}

const (
	unauthorized = "401 Unauthorized"
	oktaProvider = "okta"
)

// NewSessionHandler returns a SessionHandler with the specified URL
func NewSessionHandler(url string) (*SessionHandler, error) {
	sh := &SessionHandler{
		url: url,
	}
	return sh, nil
}

// WithAuthentication initializes the SessionHandler auth field
func (s *SessionHandler) WithAuthentication(url, providerID string, b login.BrowserI) *SessionHandler {
	s.auth = &Auth{
		browser:    b,
		providerID: providerID,
		url:        url,
	}
	return s
}

// WithBody sets the SessionHandler request body
func (s *SessionHandler) WithBody(body io.Reader) *SessionHandler {
	s.body = body
	return s
}

// Get sets the SessionHandler method to HTTP GET
func (s *SessionHandler) Get() *SessionHandler {
	s.method = "GET"
	return s
}

// Post sets the SessionHandler method to HTTP POST
func (s *SessionHandler) Post(body io.Reader) *SessionHandler {
	s.method = "POST"
	s.WithBody(body)
	return s
}

// WithClient sets the SessionHandler HTTP client
func (s *SessionHandler) WithClient(c *http.Client) *SessionHandler {
	s.client = c
	return s
}

// WithContext sets the SessionHandler miactl context
func (s *SessionHandler) WithContext(ctx string) *SessionHandler {
	s.context = ctx
	return s
}

func (s *SessionHandler) WithUrl(url string) *SessionHandler {
	s.url = url
	return s
}

// GetContext returns the SessionHandler miactl context
func (s *SessionHandler) GetContext() string {
	return s.context
}

// HTTPClientBuilder creates an HTTP client from the given CLI options
func HTTPClientBuilder(opts *clioptions.CLIOptions) (*http.Client, error) {
	client := &http.Client{}
	// TODO: extract CA certificate from viper config file
	if opts.CACert != "" || opts.Insecure {
		transport, err := configureTransport(opts)
		if err != nil {
			return nil, fmt.Errorf("error creating custom transport: %w", err)
		}
		client.Transport = transport
	}

	return client, nil
}

// ExecuteRequest executes the HTTP request with the info in the SessionHandler object.
func (s *SessionHandler) ExecuteRequest() (*http.Response, error) {
	httpReq, err := http.NewRequest(s.method, s.url, s.body)
	if err != nil {
		return nil, fmt.Errorf("error building the http request: %w", err)
	}
	token, err := s.auth.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("error retrieving token: %w", err)
	}
	q := httpReq.URL.Query()
	q.Add("environment", "test")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("client-key", "miactl")
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending the http request: %w", err)
	}
	if resp.Status == unauthorized {
		newToken, err := s.auth.Authenticate()
		if err != nil {
			return resp, fmt.Errorf("error refreshing token: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+newToken)
		resp, err = s.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("error resending the http request: %w", err)
		}
		if resp.Status == unauthorized {
			return nil, fmt.Errorf("unable to login: %s", resp.Status)
		}
	}
	return resp, nil
}

// configureTransport configures an HTTP Transport from the CLI options
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
	if opts.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	// set the TLS configuration to the HTTP transport
	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

// ParseResponseBody reads and unmarshals the response body in the given interface
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

func FakeSessionHandler(url string) *SessionHandler {
	return &SessionHandler{
		url:     url,
		context: "fake-ctx",
		client:  &http.Client{},
		auth:    &testutils.MockValidToken{},
	}
}
