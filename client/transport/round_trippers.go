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

package transport

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	netutil "github.com/mia-platform/miactl/client/net/util"
)

// WrappedRoundTripperForConfig wraps a http.RoundTripper with any relevant behavior from the config.
// You can use this method if you need to setup customize the connection but need the HTTP-like behavior.
// You will normally need only the RoundTripper returned by the New function.
func WrappedRoundTripperForConfig(config *Config, rt http.RoundTripper) http.RoundTripper {
	if config.wrapTransport != nil {
		rt = config.wrapTransport(rt)
	}

	if config.EnableDebug {
		rt = NewDebugRoundTripper(rt)
	}

	if len(config.UserAgent) > 0 {
		rt = NewUserAgentRoundTripper(config.UserAgent, rt)
	}
	return rt
}

type debugRoundTripper struct {
	next http.RoundTripper
}

// NewDebugRoundTripper allows to display in the logs output debug information on the API requests performed
// by the client. It will use the logr.Logger found in the request context.
// Well known headers that contains sensible data will be redacted.
func NewDebugRoundTripper(next http.RoundTripper) http.RoundTripper {
	return &debugRoundTripper{next: next}
}

func (rt *debugRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := netutil.CloneRequest(req)
	logger := logr.FromContextOrDiscard(req.Context())

	logger.V(6).Info(fmt.Sprintf("%s, %s", req.Method, req.URL.String()))
	logger.V(7).Info("Request Headers:")
	for headerKey, headerValues := range req.Header {
		for _, value := range headerValues {
			maskedValue := maskSensibleHeaderValue(headerKey, value)
			logger.V(7).Info(fmt.Sprintf("\t%s: %s", headerKey, maskedValue))
		}
	}

	logger.V(10).Info(fmt.Sprintf("Try this at home:\n%s", printCurl(req)))
	requestStartTime := time.Now()
	response, err := rt.next.RoundTrip(clonedReq)
	requestEndTime := time.Since(requestStartTime)
	logger.V(6).Info(fmt.Sprintf("Response Status: %s in %d milliseconds", response.Status, requestEndTime.Milliseconds()))
	logger.V(7).Info("Response Headers:")
	for headerKey, headerValues := range response.Header {
		for _, value := range headerValues {
			maskedValue := maskSensibleHeaderValue(headerKey, value)
			logger.V(7).Info(fmt.Sprintf("\t%s: %s", headerKey, maskedValue))
		}
	}
	return response, err
}

func maskSensibleHeaderValue(headerKey string, value string) string {
	// mask value only if the header is "Authorization"
	if !strings.EqualFold(headerKey, "Authorization") {
		return value
	}

	// don't do anything if the value is empty
	if len(value) == 0 {
		return ""
	}

	var authType string
	if i := strings.Index(value, " "); i > 0 {
		authType = value[0:i]
	} else {
		authType = value
	}

	switch strings.ToLower(authType) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#authentication_schemes
	case "basic", "bearer", "digest", "negotiate":
		if len(value) > len(authType)+1 {
			value = authType + " REDACTED"
		} else {
			value = authType
		}
		return value
	default:
		return "REDACTED"
	}
}

func printCurl(r *http.Request) string {
	headers := ""
	for key, values := range r.Header {
		for _, value := range values {
			value = maskSensibleHeaderValue(key, value)
			headers += fmt.Sprintf("\t-H %q\n", fmt.Sprintf("%s: %s", key, value))
		}
	}

	return fmt.Sprintf("curl -v -X%s\n%s\t'%s'", r.Method, headers, r.URL.String())
}

type userAgentRoundTripper struct {
	userAgent string
	next      http.RoundTripper
}

// NewUserAgentRoundTripper will add userAgent to the request headers unless it has already been set.
func NewUserAgentRoundTripper(userAgent string, next http.RoundTripper) http.RoundTripper {
	return &userAgentRoundTripper{
		userAgent: userAgent,
		next:      next,
	}
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("User-Agent")) != 0 {
		return rt.next.RoundTrip(req)
	}
	clonedReq := netutil.CloneRequest(req)
	clonedReq.Header.Set("User-Agent", rt.userAgent)
	return rt.next.RoundTrip(clonedReq)
}
