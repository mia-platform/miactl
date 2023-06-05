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

package cliconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachePath(t *testing.T) {
	testXDGConfig := t.TempDir()
	testXDGCache := t.TempDir()
	testCases := map[string]struct {
		expectedConfig      string
		expectedCredentials string
		expectedCache       string
		xdgConfigValue      string
		xdgCacheValue       string
		emptyHome           bool
	}{
		"Test Empty XDG environments": {
			expectedConfig:      filepath.Join(os.Getenv("HOME"), ".config", miactlFolderName, configFileName),
			expectedCache:       filepath.Join(os.Getenv("HOME"), ".cache", miactlFolderName),
			expectedCredentials: filepath.Join(os.Getenv("HOME"), ".config", miactlFolderName, credentials),
		},
		"Test Empty HOME": {
			expectedConfig:      filepath.Join("/", ".config", miactlFolderName, configFileName),
			expectedCache:       filepath.Join("/", ".cache", miactlFolderName),
			expectedCredentials: filepath.Join("/", ".config", miactlFolderName, credentials),
			emptyHome:           true,
		},
		"Test Empty HOME with XDG environments": {
			expectedConfig:      filepath.Join(testXDGConfig, miactlFolderName, configFileName),
			expectedCredentials: filepath.Join(testXDGConfig, miactlFolderName, credentials),
			expectedCache:       filepath.Join(testXDGCache, miactlFolderName),
			emptyHome:           true,
			xdgConfigValue:      testXDGConfig,
			xdgCacheValue:       testXDGCache,
		},
		"Test XDG environments": {
			expectedConfig:      filepath.Join(testXDGConfig, miactlFolderName, configFileName),
			expectedCredentials: filepath.Join(testXDGConfig, miactlFolderName, credentials),
			expectedCache:       filepath.Join(testXDGCache, miactlFolderName),
			xdgConfigValue:      testXDGConfig,
			xdgCacheValue:       testXDGCache,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			if testCase.emptyHome {
				t.Setenv("HOME", "")
			}
			t.Setenv("XDG_CONFIG_HOME", testCase.xdgConfigValue)
			t.Setenv("XDG_CACHE_HOME", testCase.xdgCacheValue)

			assert.Equal(t, testCase.expectedConfig, ConfigFilePath())
			assert.Equal(t, testCase.expectedCredentials, CredentialsFilePath())
			assert.Equal(t, testCase.expectedCache, CacheFolderPath())
		})
	}
}
