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
	"strings"
)

// defaultContentType is the default Accept and body type value
const defaultContentType = "application/json"

// Interface captures the set of operations for interacting with REST apis
type Interface interface {
	Delete() *Request
	Get() *Request
	Post() *Request
	Patch() *Request
	HTTPClient() *http.Client
}

// APIClient wrap an http.Client that can connect to Mia-Platform Console
type APIClient struct {
	baseURL       *url.URL
	contentConfig contentConfig

	client *http.Client
}

// newAPIClient create a new APIClient for url using config and httpClient for configure it
func newAPIClient(url *url.URL, config contentConfig, httpClient *http.Client) *APIClient {
	// be sure to have a valid ContetType
	if len(config.ContentType) == 0 {
		config.ContentType = defaultContentType
	}

	// normalize url
	baseURL := *url
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	return &APIClient{
		baseURL:       &baseURL,
		contentConfig: config,

		client: httpClient,
	}
}

// Get return a new Request object for a GET http request
func (c *APIClient) Get() *Request {
	return NewRequest(c).SetVerb(http.MethodGet)
}

// Post return a new Request object for a POST http request
func (c *APIClient) Post() *Request {
	return NewRequest(c).SetVerb(http.MethodPost)
}

// Delete return a new Request object for a DELETE http request
func (c *APIClient) Delete() *Request {
	return NewRequest(c).SetVerb(http.MethodDelete)
}

func (c *APIClient) Patch() *Request {
	return NewRequest(c).SetVerb(http.MethodPatch)
}

func (c *APIClient) HTTPClient() *http.Client {
	return c.client
}
