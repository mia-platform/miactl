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

package extensions

import (
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/stretchr/testify/require"
)

func TestNewActivationScope(t *testing.T) {
	testCases := map[string]struct {
		companyID     string
		projectID     string
		expectedScope ActivationScope
	}{
		"project scope": {
			companyID:     "c1",
			projectID:     "p1",
			expectedScope: ActivationScope{ContextID: "p1", ContextType: ProjectContext},
		},
		"company scope": {
			companyID:     "c1",
			expectedScope: ActivationScope{ContextID: "c1", ContextType: CompanyContext},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			scope := NewActivationScope(&client.Config{
				CompanyID: test.companyID,
				ProjectID: test.projectID,
			})
			require.Equal(t, test.expectedScope, scope)
		})
	}
}

func TestActivationScopeString(t *testing.T) {
	scope := ActivationScope{
		ContextID:   "c1",
		ContextType: CompanyContext,
	}
	t.Run("in a formatted string", func(t *testing.T) {
		require.Equal(t, ">>company: c1<<", fmt.Sprintf(">>%s<<", scope))
	})

	t.Run("String", func(t *testing.T) {
		require.Equal(t, "company: c1", scope.String())
	})
}
