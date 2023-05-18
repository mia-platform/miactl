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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Request wrap the http.Request configuration providing functions for configure it in an easier and contained way
type Request struct {
	restClient *APIClient

	verb    string
	apiPath string
	params  url.Values
	headers http.Header

	err  error
	body []byte
}

// NewRequest creates a new request helper object for calling Mia-Platform Console API
func NewRequest(client *APIClient) *Request {
	request := &Request{
		restClient: client,
	}

	switch {
	case len(client.contentConfig.AcceptContentTypes) > 0:
		request.SetHeader("Accept", client.contentConfig.AcceptContentTypes)
	case len(client.contentConfig.ContentType) > 0:
		request.SetHeader("Accept", client.contentConfig.ContentType+", */*")
	}
	return request
}

// SetHeader add the passed value for the header key, if the key aldready exists, it will be removed
func (r *Request) SetHeader(key, value string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}

	r.headers.Del(key)
	r.headers.Add(key, value)
	return r
}

// SetVerb set the HTTP verb for the request
func (r *Request) SetVerb(verb string) *Request {
	r.verb = verb
	return r
}

// SetParam set the HTTP query params key to value if no error has been found in the request construction
func (r *Request) SetParam(key string, values ...string) *Request {
	if r.err != nil {
		return r
	}

	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[key] = append(r.params[key], values...)
	return r
}

// APIPath set the apiPath for the request, if apiPath cannot be parse as a valid path an error can be
// found with the Error function. It will be ensured that the path will always end with /.
func (r *Request) APIPath(apiPath string) *Request {
	if r.err != nil {
		return r
	}

	// parse the string to be sure is a valid url path, discard all the rest
	parsedURI, err := url.Parse(apiPath)
	if err != nil {
		r.err = err
		return r
	}

	r.apiPath = parsedURI.Path
	if !strings.HasSuffix(r.apiPath, "/") {
		r.apiPath += "/"
	}
	return r
}

// Body set the body of the request if no error has been found in the request construction
func (r *Request) Body(bodyBytes []byte) *Request {
	if r.err != nil {
		return r
	}

	r.body = bodyBytes
	return r
}

// Error return any error encountered constructing the request, if any.
func (r *Request) Error() error {
	return r.err
}

// URL return the url that will be used by the http.Request in this moment
func (r *Request) URL() *url.URL {
	url := *r.restClient.baseURL

	url.Path = r.apiPath
	url.RawQuery = r.params.Encode()

	return &url
}

// preflightCheck perform checks for human error in setting the request
func (r *Request) preflightCheck() error {
	// if the request has already an error return it
	if r.err != nil {
		return r.err
	}

	// if no verb has been set break
	if len(r.verb) == 0 {
		return fmt.Errorf("no HTTP verb specified on request")
	}

	switch {
	case r.verb == http.MethodPost && len(r.body) == 0:
		return fmt.Errorf("empty body for a POST request")
	case r.verb == http.MethodGet && len(r.body) > 0:
		return fmt.Errorf("body set for a GET request")
	}

	return nil
}

// Do execute the http.Request in the provided context
func (r *Request) Do(ctx context.Context) (*Response, error) {
	var response *Response
	err := r.request(ctx, func(req *http.Request, res *http.Response) {
		response = parseRequestAndResponse(req, res)
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	if err := r.preflightCheck(); err != nil {
		return err
	}

	httpRequest, err := r.httpRequest(ctx)
	if err != nil {
		return err
	}

	response, err := r.restClient.client.Do(httpRequest)
	func() {
		if response != nil {
			defer response.Body.Close()
			fn(httpRequest, response)
		}
	}()
	return err
}

// httpRequest create a new http.Request from Request or an error
func (r *Request) httpRequest(ctx context.Context) (*http.Request, error) {
	bodyReader := bytes.NewReader(r.body)
	url := r.URL().String()
	req, err := http.NewRequest(r.verb, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header = r.headers
	return req, nil
}

func parseRequestAndResponse(req *http.Request, resp *http.Response) *Response {
	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return &Response{
				rawResponse: resp,
				rawRequest:  req,
				err:         fmt.Errorf("unexpected error when reading body: %w", err),
			}
		}
		body = data
	}

	if resp.StatusCode > http.StatusBadRequest {
		return &Response{
			rawResponse: resp,
			rawRequest:  req,
			statusCode:  resp.StatusCode,
			err:         &ResponseError{body: body},
		}
	}

	return &Response{
		rawResponse: resp,
		rawRequest:  req,
		body:        body,
		statusCode:  resp.StatusCode,
	}
}
