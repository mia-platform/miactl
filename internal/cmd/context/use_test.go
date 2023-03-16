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

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const missingContextName = "missing"
const contextName1 = "context1"

func TestContextLookUpEmptyContextMap(t *testing.T) {
	viper.SetConfigType("yaml")
	config := ``
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	context, err := contextLookUp(missingContextName)
	require.Nil(t, context)
	require.EqualError(t, err, "no context specified in config file")
}

func TestContextLookUp(t *testing.T) {
	viper.SetConfigType("yaml")
	config := `contexts:
  context1:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	context, err := contextLookUp(missingContextName)
	require.Nil(t, context)
	require.EqualError(t, err, "context missing does not exist")

	context, err = contextLookUp(contextName1)
	require.Nil(t, err)
	require.NotNil(t, context)
	require.Equal(t, "http://url", context["apibaseurl"])
	require.Equal(t, "123", context["companyid"])
	require.Equal(t, "123", context["projectid"])
}
