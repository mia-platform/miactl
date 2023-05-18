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

	"github.com/mia-platform/miactl/internal/transport"
)

// httpClientForConfig return a new http.Client with the transport security provided in the config
// Will return the default http.DefaultClient if no special case behavior is needed.
func httpClientForConfig(config *Config) (*http.Client, error) {
	httpClient := http.DefaultClient
	transportConfig := &transport.Config{
		UserAgent: config.UserAgent,
		TLSConfig: transport.TLSConfig{
			Insecure: config.Insecure,
			CAFile:   config.CAFile,
		},
	}

	transport, err := transport.NewTransport(transportConfig)
	if err != nil {
		return nil, err
	}

	if transport != http.DefaultTransport || config.Timeout > 0 {
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		}
	}

	return httpClient, nil
}
