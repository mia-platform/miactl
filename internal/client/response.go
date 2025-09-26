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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mia-platform/miactl/internal/resources"
)

// ResponseError represent an error from an api call
type ResponseError struct {
	body       []byte
	statusCode int
}

// Error return the message of the api call error
func (r *ResponseError) Error() string {
	var out *resources.APIError
	err := parseBody(r.body, &out)
	if err != nil {
		return fmt.Sprintf("cannot parse server response: %s", err)
	}

	if len(out.Message) > 0 {
		return out.Message
	}

	switch r.statusCode {
	case http.StatusBadRequest:
		return "something went wrong"
	case http.StatusForbidden:
		return "you are not allowed to make this call, contact your admin"
	default:
		return fmt.Sprintf("error received from remote: %d", r.statusCode)
	}
}

// Response wrap an http.Response and provide functions to operate safely on it
type Response struct {
	rawResponse *http.Response
	rawRequest  *http.Request

	body       []byte
	statusCode int
	err        error
}

// RawRequest return the original request if you need to access things not exposed by the other functions
func (r *Response) RawRequest() *http.Request {
	return r.rawRequest
}

// StatusCode return the response status code
func (r *Response) StatusCode() int {
	return r.statusCode
}

// Error return the error found in the response
func (r *Response) Error() error {
	switch {
	case r.err != nil:
		return r.err
	case r.statusCode >= http.StatusBadRequest:
		return &ResponseError{body: r.body, statusCode: r.statusCode}
	default:
		return nil
	}
}

// ParseResponse will parse the underlying body inside the obj passed
func (r *Response) ParseResponse(obj interface{}) error {
	if r.err != nil {
		return r.err
	}

	return parseBody(r.body, obj)
}

func parseBody(body []byte, obj interface{}) error {
	err := json.Unmarshal(body, obj)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("cannot parse server response: %w", err)
	}

	return nil
}
