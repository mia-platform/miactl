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

// RESTClientForConfig create a RESTClient with config
func RESTClientForConfig(config *Config) (*RESTClient, error) {
	// Validate Host before constructing the transport/client so we can fail fast.
	if _, err := defaultServerURL(config); err != nil {
		return nil, err
	}

	httpClient, err := httpClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return restClientForConfigAndClient(config, httpClient)
}

// restClientForConfigAndClient create a RESTClient with config and httpClient
func restClientForConfigAndClient(config *Config, httpClient *http.Client) (*RESTClient, error) {
	baseURL, err := defaultServerURL(config)
	if err != nil {
		return nil, err
	}

	contentConfig := contentConfig{
		AcceptContentTypes: "application/json",
		ContentType:        "application/json",
	}
	return newRESTClient(baseURL, contentConfig, httpClient), nil
}
