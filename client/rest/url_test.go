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

package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultServerUrl(t *testing.T) {
	testCases := []struct {
		Host        string
		ExpectedURL string
		ExpectedErr bool
	}{
		{"", "", true},
		{"127.0.0.1", "https://127.0.0.1/", false},
		{"127.0.0.1:8080", "https://127.0.0.1:8080/", false},
		{"foo.bar.com", "https://foo.bar.com/", false},
		{"http://host/prefix", "http://host/prefix", false},
		{"http://host/", "http://host/", false},
		{"host/server", "", true},
	}

	for _, testCase := range testCases {
		u, err := defaultServerURL(&Config{Host: testCase.Host})
		if testCase.ExpectedErr {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, testCase.ExpectedURL, u.String())
	}
}
