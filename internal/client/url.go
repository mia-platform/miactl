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
	"fmt"
	"net/url"
	"path"
)

// defaultServerURL is used for validating the Host and return a complete url, a sanitized APIPath or an err
func defaultServerURL(config *Config) (*url.URL, error) {
	host := config.Host
	if host == "" {
		return nil, fmt.Errorf("host must be a URL or a host:port pair")
	}

	hostURL, err := url.Parse(host)
	if err != nil || hostURL.Scheme == "" || hostURL.Host == "" {
		// always assume an https url if no schema is passed
		hostURL, err = url.Parse("https://" + host)
		if err != nil {
			return nil, err
		}
		if hostURL.Path != "" && hostURL.Path != "/" {
			return nil, fmt.Errorf("host must be a URL or a host:port pair: %q", host)
		}
	}

	// made sure the hostURL end with / if is the root
	hostURL.Path = path.Join(hostURL.Path, "/")
	return hostURL, nil
}
