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
	Get() *Request
	Post() *Request
}

// RESTClient wrap an http.Client that can connect to Mia-Platform Console
type RESTClient struct {
	baseURL       *url.URL
	contentConfig contentConfig

	client *http.Client
}

// newRESTClient create a new RESTClient for url using config and httpClient for configure it
func newRESTClient(url *url.URL, config contentConfig, httpClient *http.Client) *RESTClient {
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

	return &RESTClient{
		baseURL:       &baseURL,
		contentConfig: config,

		client: httpClient,
	}
}

// Get return a new Request object for a GET http request
func (c *RESTClient) Get() *Request {
	return NewRequest(c).SetVerb(http.MethodGet)
}

// Get return a new Request object for a POST http request
func (c *RESTClient) Post() *Request {
	return NewRequest(c).SetVerb(http.MethodPost)
}
