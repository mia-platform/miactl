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

// Config transport layer configurations for setting up http.Transport
type Config struct {
	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string
	// TLSConfig contains settings to enable transport layer security
	TLSConfig
}

// TLSConfig contains settings to enable transport layer security
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool

	// Trusted root certificates for server
	CAFile string
}

func (c *Config) HasCA() bool {
	return len(c.CAFile) > 0
}
