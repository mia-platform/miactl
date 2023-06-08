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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIClientForConfig(t *testing.T) {
	testCases := map[string]struct {
		Config *Config
		Err    bool
	}{
		"good REST client": {
			Config: &Config{
				Host: "https://test",
			},
		},
		"fail fast with wrong host": {
			Config: &Config{
				Host: "host/server",
			},
			Err: true,
		},
		"fail for error in transport": {
			Config: &Config{
				Host: "http://test",
				TLSClientConfig: TLSClientConfig{
					CAFile: "invalid.file",
				},
			},
			Err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			restClient, err := APIClientForConfig(testCase.Config)
			switch {
			case testCase.Err:
				assert.Error(t, err)
				assert.Nil(t, restClient)
			case !testCase.Err:
				assert.NoError(t, err)
				assert.NotNil(t, restClient)
			}
		})
	}
}
