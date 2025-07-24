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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRef(t *testing.T) {
	tests := []struct {
		refType          string
		refName          string
		expectError      bool
		expectedErrorMsg string
	}{
		{"revisions", "rev1", false, ""},
		{"versions", "v1", false, ""},
		{"branches", "branch1", false, ""},
		{"tags", "tag1", false, ""},
		{"revisions", "with/slash", false, ""},
		{"versions", "with/slash", false, ""},
		{"branches", "with/slash", false, ""},
		{"tags", "with/slash", false, ""},

		{"invalidType", "name", true, "unknown reference type: invalidType"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.refType, tt.refName), func(t *testing.T) {
			ref, err := NewRef(tt.refType, tt.refName)
			if tt.expectError {
				require.EqualError(t, err, tt.expectedErrorMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.refType, ref.refType)
			require.Equal(t, tt.refName, ref.refName)
		})
	}
}

func TestGetEncodedResourceLocation(t *testing.T) {
	tests := []struct {
		refType          string
		refName          string
		expected         string
		expectError      bool
		expectedErrorMsg string
	}{
		{"revisions", "rev1", "revisions/rev1", false, ""},
		{"versions", "v1", "versions/v1", false, ""},
		{"branches", "branch1", "branches/branch1", false, ""},
		{"tags", "tag1", "branches/tag1", false, ""},
		{"revisions", "with/slash", "revisions/with%2Fslash", false, ""},
		{"versions", "with/slash", "versions/with%2Fslash", false, ""},
		{"branches", "with/slash", "branches/with%2Fslash", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			ref, err := NewRef(tt.refType, tt.refName)
			require.NoError(t, err)

			encodedString := ref.EncodedLocationPath()
			require.Equal(t, encodedString, tt.expected)
		})
	}
}
