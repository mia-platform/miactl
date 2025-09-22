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

package client

import (
	"net/http"
	"net/url"
	"time"
)

// Config holds the common attributes that can be passed to a client on initialization
type Config struct {
	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig
	// AuthConfig contains settings for settign up authentication for the http requests
	AuthConfig
	// AuthCacheReadWriter provides access to authorization cache
	AuthCacheReadWriter

	// Host must be a host string, a host:port pair, or a URL to the base of the server.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the server. This allows a frontend
	// proxy to easily relocate all of the server endpoints.
	Host string

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// Transport add a custom transport instead of creating a new one. Wrappers will be added to it
	Transport http.RoundTripper

	// CompanyID contains the company id that can be used for filtering requests
	CompanyID string

	// ProjectID contains the project id that can be used for filtering requests
	ProjectID string

	// Environment contains the environment scope that can be used for filtering requests
	Environment string

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool

	// Trusted root certificates for server
	CAFile string
}

// AuthConfig contains settings for settign up authentication for the http requests
type AuthConfig struct {
	ClientID          string
	ClientSecret      string
	JWTKeyID          string
	JWTPrivateKeyData string
}

// contentConfig contains settings that affect how objects are transformed when sent to the server.
type contentConfig struct {
	// AcceptContentTypes specifies the types the client will accept and is optional.
	// If not set, ContentType will be used to define the Accept header
	AcceptContentTypes string
	// ContentType specifies the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string
}

// APIClientForConfig create a APIClient with config
func APIClientForConfig(config *Config) (*APIClient, error) {
	// Validate Host before constructing the transport/client so we can fail fast.
	baseURL, err := defaultServerURL(config)
	if err != nil {
		return nil, err
	}

	httpClient, err := httpClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return apiClientForHTTPClient(baseURL, httpClient)
}

// apiClientForConfigAndClient create a APIClient with config and httpClient
func apiClientForHTTPClient(baseURL *url.URL, httpClient *http.Client) (*APIClient, error) {
	contentConfig := contentConfig{
		AcceptContentTypes: "application/json",
		ContentType:        "application/json",
	}
	return newAPIClient(baseURL, contentConfig, httpClient), nil
}
