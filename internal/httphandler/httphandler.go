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
	"fmt"
	"io"
	"net/http"

	"github.com/mia-platform/miactl/old/sdk"
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

func RequestBuilder(opts sdk.Options, authFn Authenticate) *Request {
	req := &Request{
		url:    opts.APIBaseURL,
		client: &http.Client{},
		authFn: authFn,
	}
	return req
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
