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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/mia-platform/miactl/internal/cliconfig"
)

func TestRemoveContext(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(wd, "testdata", "remove-config.yaml")

	testCases := map[string]struct {
		contextName       string
		expectErr         bool
		expectAuthRemains string
		expectAuthGone    string
		expectNoCurrent   bool
	}{
		"context not found": {
			contextName: "does-not-exist",
			expectErr:   true,
		},
		"auth still referenced by another context survives": {
			contextName:       "ctx-shared-a",
			expectAuthRemains: "shared-auth",
		},
		"last context referencing an auth removes it": {
			contextName:    "ctx-solo",
			expectAuthGone: "solo-auth",
		},
		"removing the current context clears it": {
			contextName:     "ctx-solo",
			expectAuthGone:  "solo-auth",
			expectNoCurrent: true,
		},
		"context without a linked auth has no side effects": {
			contextName: "ctx-no-auth",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			tempFile := copyFile(t, testdata)
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = tempFile

			err := removeContext(testCase.contextName, locator)
			if testCase.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			config, err := locator.ReadConfig()
			require.NoError(t, err)

			_, found := config.Contexts[testCase.contextName]
			assert.False(t, found, "context should have been removed")

			if testCase.expectAuthRemains != "" {
				_, found := config.Auth[testCase.expectAuthRemains]
				assert.True(t, found, "auth still referenced by another context should remain")
			}
			if testCase.expectAuthGone != "" {
				_, found := config.Auth[testCase.expectAuthGone]
				assert.False(t, found, "orphaned auth should have been removed")
			}
			if testCase.expectNoCurrent {
				assert.Empty(t, config.CurrentContext)
			}
		})
	}
}

func TestRemoveContextDeletesCachedToken(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdata := filepath.Join(wd, "testdata", "remove-config.yaml")
	tempFile := copyFile(t, testdata)

	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	locator := cliconfig.NewConfigPathLocator()
	locator.ExplicitPath = tempFile
	config, err := locator.ReadConfig()
	require.NoError(t, err)

	ctxConfig := config.Contexts["ctx-solo"]
	authConfig := config.Auth["solo-auth"]
	rw := cliconfig.NewAuthReadWriter(locator, ctxConfig, authConfig)
	rw.WriteJWTToken(&oauth2.Token{AccessToken: "test-token"})

	entriesBefore, err := os.ReadDir(filepath.Join(cacheDir, "miactl"))
	require.NoError(t, err)
	require.Len(t, entriesBefore, 1)

	require.NoError(t, removeContext("ctx-solo", locator))

	entriesAfter, err := os.ReadDir(filepath.Join(cacheDir, "miactl"))
	require.NoError(t, err)
	assert.Empty(t, entriesAfter)
}
