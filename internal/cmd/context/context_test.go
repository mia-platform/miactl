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

func TestNewContextCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewContextCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestGetContextBaseURL(t *testing.T) {
	viper.SetConfigType("yaml")

	// valid config1 and get
	config1 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err := viper.ReadConfig(strings.NewReader(config1))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	// valid context
	url, err := GetContextBaseURL("fake-ctx")
	require.NoError(t, err)
	require.Equal(t, "http://url", url)

	// wrong context name
	url, err = GetContextBaseURL("wrong-ctx")
	require.Error(t, err)
	require.Equal(t, "", url)
}

func TestGetContextCompanyID(t *testing.T) {
	viper.SetConfigType("yaml")

	config1 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err := viper.ReadConfig(strings.NewReader(config1))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	// valid context
	companyID, err := GetContextCompanyID("fake-ctx")
	require.NoError(t, err)
	require.Equal(t, "123", companyID)

	// wrong context name
	companyID, err = GetContextCompanyID("wrong-ctx")
	require.Error(t, err)
	require.Equal(t, "", companyID)

	// company id unset
	config2 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err = viper.ReadConfig(strings.NewReader(config2))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	companyID, err = GetContextCompanyID("wrong-ctx")
	require.Error(t, err)
	require.Equal(t, "", companyID)
}

func TestGetCurrentContext(t *testing.T) {
	viper.SetConfigType("yaml")

	// current context set
	config1 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"
current-context: fake-ctx`
	err := viper.ReadConfig(strings.NewReader(config1))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	currentContext, err := GetCurrentContext()
	require.NoError(t, err)
	require.Equal(t, "fake-ctx", currentContext)

	// current context unset
	config2 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err = viper.ReadConfig(strings.NewReader(config2))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	currentContext, err = GetCurrentContext()
	require.Error(t, err)
	require.Equal(t, "", currentContext)
}
