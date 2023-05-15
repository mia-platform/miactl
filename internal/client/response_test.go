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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResponse(t *testing.T) {
	testCases := map[string]struct {
		response *Response
		err      bool
	}{
		"response with body": {
			response: &Response{body: []byte(`{"key": "value"}`)},
		},
		"response with different interface": {
			response: &Response{body: []byte(`[{"key": "value"}]`)},
			err:      true,
		},
		"response with error": {
			response: &Response{err: fmt.Errorf("network error")},
			err:      true,
		},
	}

	type testObj struct {
		Key string
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			var out *testObj
			err := testCase.response.ParseResponse(&out)
			if testCase.err {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	}
}
