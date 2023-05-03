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

package context

import (
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewListContextsCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewListContextsCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestGetContextNames(t *testing.T) {
	viper.SetConfigType("yaml")

	testCases := []struct {
		name        string
		config      string
		expectedOut []string
		expectedErr string
	}{
		{
			name:        "valid config and context",
			config:      valid,
			expectedOut: []string{"fake-ctx", "other-ctx"},
			expectedErr: "",
		},
		{
			name:        "valid config and context",
			config:      valid,
			expectedOut: []string{"fake-ctx", "other-ctx"},
			expectedErr: "",
		},
	}

	for _, tc := range testCases {
		err := viper.ReadConfig(strings.NewReader(tc.config))
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}
		names, err := getContextNames()
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
		require.Equal(t, tc.expectedOut, names)
	}
}
