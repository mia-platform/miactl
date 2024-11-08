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

package rest

import (
	"net/http"
	"time"
)

// Config holds the common attributes that can be passed to a client on initialization
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the server.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the server. This allows a frontend
	// proxy to easily relocate all of the server endpoints.
	Host string

	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// DisableCompression bypasses automatic GZip compression requests to the
	// server.
	DisableCompression bool

	// Transport add a custom transport instead of creating a new one. Wrappers will be added to it
	Transport http.RoundTripper

	// AuthConfig contains settings for settign up authentication for the http requests
	AuthConfig *AuthConfig

	// AuthCacheReadWriter provides access to authorization cache
	AuthCacheReadWriter

	// CompanyID contains the company id that can be used for filtering requests
	CompanyID string

	// ProjectID contains the project id that can be used for filtering requests
	ProjectID string

	// Environment contains the environment scope that can be used for filtering requests
	Environment string

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration

	// EnableDebug enable connection debug information to be outputted to the logr.Logger level found in the request context.
	EnableDebug bool
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

// ClientForConfig create a APIClient with config
func ClientForConfig(config *Config) (*Client, error) {
	// Validate Host before constructing the transport/client so we can fail fast.
	_, err := defaultServerURL(config)
	if err != nil {
		return nil, err
	}

	httpClient, err := HTTPClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return ClientForConfigAndHTTPClient(config, httpClient)
}

// ClientForConfigAndHTTPClient create a Client with config and httpClient
func ClientForConfigAndHTTPClient(config *Config, httpClient *http.Client) (*Client, error) {
	baseURL, err := defaultServerURL(config)
	if err != nil {
		return nil, err
	}

	contentConfig := ContentConfig{
		AcceptContentTypes: "application/json",
		ContentType:        "application/json",
	}
	return NewRESTClient(baseURL, contentConfig, httpClient), nil
}
