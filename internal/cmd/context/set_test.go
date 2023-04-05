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

func TestUpdateContextMap(t *testing.T) {
	err := viper.ReadConfig(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	// Test creating a new context
	opts := &clioptions.CLIOptions{Endpoint: "https://url", ProjectID: "project1", CompanyID: "company1", CACert: "/path/to/cert"}
	newContext := map[string]interface{}{"endpoint": "https://url", "projectid": "project1", "companyid": "company1", "ca-cert": "/path/to/cert"}
	expectedContexts := make(map[string]interface{})
	expectedContexts["context1"] = newContext
	actualContexts := updateContextMap(opts, "context1")
	require.Equal(t, expectedContexts, actualContexts)

	// Test updating the existing context
	opts = &clioptions.CLIOptions{Endpoint: "https://url2", ProjectID: "project2", CompanyID: "company2", CACert: "/path/to/cert"}
	updatedContext := map[string]interface{}{"endpoint": "https://url2", "projectid": "project2", "companyid": "company2", "ca-cert": "/path/to/cert"}
	expectedContexts["context1"] = updatedContext
	actualContexts = updateContextMap(opts, "context1")
	require.Equal(t, expectedContexts, actualContexts)
}
