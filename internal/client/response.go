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
	"fmt"
	"io"
	"net/http"
)

// ResponseError represent an error from an api call
type ResponseError struct {
	body []byte
}

// Error return the message of the api call error
func (r *ResponseError) Error() string {
	// TODO
	return ""
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

// Error return the response status code
func (r *Response) StatusCode() int {
	return r.statusCode
}

// Error return the error found in the response
func (r *Response) Error() error {
	return r.err
}

// ParseResponse will parse the underlying body inside the obj passed
func (r *Response) ParseResponse(obj interface{}) error {
	if r.err != nil {
		return r.err
	}

	err := json.Unmarshal(r.body, obj)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error during response parsing: %w", err)
	}

	return nil
}
