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

package configuration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEncodedRef(t *testing.T) {
	tests := []struct {
		name         string
		revisionName string
		versionName  string
		expected     string
		expectError  string
	}{
		{
			name:         "both revision and version specified",
			revisionName: "rev1",
			versionName:  "v1",
			expectError:  "both revision and version specified, please provide only one",
		},
		{
			name:         "only revision specified",
			revisionName: "rev1",
			expected:     "revisions/rev1",
		},
		{
			name:        "only version specified",
			versionName: "v1",
			expected:    "versions/v1",
		},
		{
			name:        "neither revision nor version specified",
			expectError: "missing revision/version name, please provide one as argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetEncodedRef(tt.revisionName, tt.versionName)
			if tt.expectError != "" {
				require.EqualError(t, err, tt.expectError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}
