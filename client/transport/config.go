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

import "net/http"

// Config transport layer configurations for setting up http.Transport
type Config struct {
	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// TLSConfig contains settings to enable transport layer security
	TLSConfig

	// DisableCompression bypasses automatic GZip compression requests to the
	// server.
	DisableCompression bool

	// EnableDebug enable the debug midlleware for printing debug information.
	EnableDebug bool

	wrapTransport WrapperFunc
}

// HasCA returns wether the configuration has a CA associated to it or not.
func (c *Config) HasCA() bool {
	return len(c.CAFile) > 0
}

// Wrap adds a middleware function that will wrap the underling http.Trasport configured prior to the
// first call is made. The provided function will be called after any existing transport wrappers are invoked.
func (c *Config) Wrap(fn WrapperFunc) {
	c.wrapTransport = Wrappers(c.wrapTransport, fn)
}

// TLSConfig contains settings to enable transport layer security
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool

	// Trusted root certificates for server
	CAFile string
}

// WrapperFunc wraps an http.RoundTripper when a new transport is created for a client, allowing per connection
// behavior to be injected.
type WrapperFunc func(rt http.RoundTripper) http.RoundTripper

// Wrappers accept one or more WrapperFunc and returns a new one
// that is equivalent to calling each of the functions one after the other in specified order. Any nil value
// will be ignored.
func Wrappers(fns ...WrapperFunc) WrapperFunc {
	if len(fns) == 0 {
		return nil
	}

	// optimize the case where we are wrapping a nil wrapper with a single new function
	if len(fns) == 2 && fns[0] == nil {
		return fns[1]
	}

	return func(rt http.RoundTripper) http.RoundTripper {
		base := rt
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			base = fn(base)
		}

		return base
	}
}
