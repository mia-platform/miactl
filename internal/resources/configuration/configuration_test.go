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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDescribeConfiguration(t *testing.T) {
	testCases := map[string]struct {
		input            map[string]any
		expectError      bool
		expectedErrorMsg string
		expectedCharts   map[string]any
	}{
		"missing config key returns error": {
			input:            map[string]any{"name": "test-project"},
			expectError:      true,
			expectedErrorMsg: "provided configuration is not valid: 'config' key not found",
		},
		"invalid config value returns error": {
			input:            map[string]any{"config": "not-a-map"},
			expectError:      true,
			expectedErrorMsg: "provided configuration is not valid: 'config' key is not a valid map[string]any",
		},
		"charts is parsed and separated from config": {
			input: map[string]any{
				"config": map[string]any{
					"name": "test-project",
				},
				"charts": map[string]any{
					"/my-chart/Chart.yaml": map[string]any{
						"pathSegments": []any{"my-chart", "Chart.yaml"},
						"content":      "apiVersion: v2\nname: my-chart\nversion: 0.1.0\n",
					},
					"/my-chart/templates/deployment.yaml": map[string]any{
						"pathSegments": []any{"my-chart", "templates", "deployment.yaml"},
						"content":      "apiVersion: apps/v1\nkind: Deployment\n",
					},
				},
			},
			expectedCharts: map[string]any{
				"/my-chart/Chart.yaml": map[string]any{
					"pathSegments": []any{"my-chart", "Chart.yaml"},
					"content":      "apiVersion: v2\nname: my-chart\nversion: 0.1.0\n",
				},
				"/my-chart/templates/deployment.yaml": map[string]any{
					"pathSegments": []any{"my-chart", "templates", "deployment.yaml"},
					"content":      "apiVersion: apps/v1\nkind: Deployment\n",
				},
			},
		},
		"charts is absent when not present in input": {
			input: map[string]any{
				"config": map[string]any{"name": "test-project"},
			},
			expectedCharts: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			config, err := BuildDescribeConfiguration(tc.input)
			if tc.expectError {
				require.EqualError(t, err, tc.expectedErrorMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCharts, config.Charts)
			assert.NotContains(t, config.Config, "charts")
		})
	}
}

func TestBuildDescribeFromFlatConfiguration(t *testing.T) {
	testCases := map[string]struct {
		input          map[string]any
		expectedCharts map[string]any
		expectedConfig map[string]any
	}{
		"charts is extracted and removed from base config": {
			input: map[string]any{
				"name": "test-project",
				"charts": map[string]any{
					"/my-chart/Chart.yaml": map[string]any{
						"pathSegments": []any{"my-chart", "Chart.yaml"},
						"content":      "apiVersion: v2\nname: my-chart\nversion: 0.1.0\n",
					},
				},
			},
			expectedCharts: map[string]any{
				"/my-chart/Chart.yaml": map[string]any{
					"pathSegments": []any{"my-chart", "Chart.yaml"},
					"content":      "apiVersion: v2\nname: my-chart\nversion: 0.1.0\n",
				},
			},
			expectedConfig: map[string]any{"name": "test-project"},
		},
		"charts is absent when not present in input": {
			input: map[string]any{
				"name": "test-project",
			},
			expectedCharts: nil,
			expectedConfig: map[string]any{"name": "test-project"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			config, err := BuildDescribeFromFlatConfiguration(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCharts, config.Charts)
			assert.NotContains(t, config.Config, "charts")
			if tc.expectedConfig != nil {
				assert.Equal(t, tc.expectedConfig, config.Config)
			}
		})
	}
}
