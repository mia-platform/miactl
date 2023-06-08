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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClientForConfig(t *testing.T) {
	testCases := map[string]struct {
		Config  *Config
		Err     bool
		Default bool
	}{
		"default http client": {
			Default: true,
			Config:  &Config{},
		},
		"custom http client for Timeout": {
			Config: &Config{
				Timeout: 30 * time.Second,
			},
		},
		"custom http client for Insecure": {
			Config: &Config{
				TLSClientConfig: TLSClientConfig{
					Insecure: true,
				},
			},
		},
		"error if transport creation fail": {
			Err: true,
			Config: &Config{
				TLSClientConfig: TLSClientConfig{
					CAFile: "invalid file",
				},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			httpClient, err := httpClientForConfig(testCase.Config)
			if testCase.Err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			switch {
			case testCase.Default:
				assert.Equal(t, http.DefaultClient, httpClient)
			case !testCase.Default:
				assert.NotEqual(t, http.DefaultClient, httpClient)
			}
		})
	}
}
