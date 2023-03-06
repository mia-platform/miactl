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
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
)

type Request struct {
	url    string
	method string
	body   io.ReadCloser
	client *http.Client
	authFn Authenticate
}

const unauthorized = "401"

type Authenticate func(url string) (string, error)

func (r *Request) WithBody(body io.ReadCloser) *Request {
	r.body = body
	return r
}

func (r *Request) Get() *Request {
	r.method = "GET"
	return r
}

func (r *Request) Post(body io.ReadCloser) *Request {
	r.method = "POST"
	r.WithBody(body)
	return r
}

func RequestBuilder(opts *clioptions.CLIOptions, authFn Authenticate) (*Request, error) {
	var client *http.Client
	if opts.CACert != "" {
		transport, err := createCustomTransport(opts.CACert)
		if err != nil {
			return nil, fmt.Errorf("error creating custom transport: %w", err)
		}
		client = &http.Client{
			Transport: transport,
		}
	}
	req := &Request{
		url:    opts.APIBaseURL,
		client: client,
		authFn: authFn,
	}
	return req, nil
}

func (r *Request) Execute() (*http.Response, error) {
	httpReq, err := http.NewRequest(r.method, r.url, r.body)
	if err != nil {
		return nil, fmt.Errorf("error building the http request: %w", err)
	}
	token, err := r.authFn(r.url)
	if err != nil {
		return nil, fmt.Errorf("error retrieving token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	resp, err := r.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending the http request: %w", err)
	}
	if resp.Status == unauthorized {
		newToken, err := r.authFn(r.url)
		if err != nil {
			return resp, fmt.Errorf("error refreshing token: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+newToken)
		resp, err = r.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("error resending the http request: %w", err)
		}
	}
	return resp, nil
}

func createCustomTransport(caCertPath string) (*http.Transport, error) {
	var rootCAs *x509.CertPool
	// load the contents of the CA certificate file
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate from path %s: %w", caCertPath, err)
	}

	// create a new CertPool object and parse the CA certificate
	rootCAs = x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// create a new TLS configuration object and set the root CAs
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	// create a new HTTP transport object and set the TLS configuration
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return transport, nil
}
