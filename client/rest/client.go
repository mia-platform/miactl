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
	"net/url"
	"strings"
)

// defaultContentType is the default Accept and body type value
const defaultContentType = "application/json"

// Interface captures the set of operations for interacting with REST apis
type Interface interface {
	Verb(verb string) *Request
	Get() *Request
	Post() *Request
	Patch() *Request
	Delete() *Request
	HTTPClient() *http.Client
}

// ContentConfig contains settings that affect how objects are transformed when sent to the server.
type ContentConfig struct {
	// AcceptContentTypes specifies the types the client will accept and is optional.
	// If not set, ContentType will be used to define the Accept header
	AcceptContentTypes string
	// ContentType specifies the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string
}

type Client struct {
	baseURL       *url.URL
	contentConfig ContentConfig

	client *http.Client
}

// NewRESTClient create a new Client for url using config and httpClient for configure it
func NewRESTClient(url *url.URL, config ContentConfig, httpClient *http.Client) *Client {
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

	return &Client{
		baseURL:       &baseURL,
		contentConfig: config,

		client: httpClient,
	}
}

// Verb return a new Request object for the given Verb
func (c *Client) Verb(verb string) *Request {
	return NewRequest(c).SetVerb(verb)
}

// Get return a new Request object for a GET http request
func (c *Client) Get() *Request {
	return NewRequest(c).SetVerb(http.MethodGet)
}

// Post return a new Request object for a POST http request
func (c *Client) Post() *Request {
	return NewRequest(c).SetVerb(http.MethodPost)
}

// Put return a new Request object for a Put http request
func (c *Client) Put() *Request {
	return NewRequest(c).SetVerb(http.MethodPut)
}

// Delete return a new Request object for a DELETE http request
func (c *Client) Delete() *Request {
	return NewRequest(c).SetVerb(http.MethodDelete)
}

func (c *Client) Patch() *Request {
	return NewRequest(c).SetVerb(http.MethodPatch)
}

func (c *Client) HTTPClient() *http.Client {
	return c.client
}
