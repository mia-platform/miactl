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

package serviceaccount

import (
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/require"
)

var (
	opts1 = clioptions.CLIOptions{
		Endpoint:  "http://url",
		CompanyID: "123",
		ProjectID: "123",
	}
	opts2 = clioptions.CLIOptions{
		Endpoint:  "http://url",
		ProjectID: "123",
	}
)

func TestNewCreateServiceAccountCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewCreateServiceAccountCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestCheckCompanyRole(t *testing.T) {
	validRoles := []string{
		"company-owner",
		"project-admin",
		"maintainer",
		"developer",
		"reporter",
		"guest",
	}
	for _, role := range validRoles {
		err := checkCompanyRole(role)
		require.NoError(t, err)
	}
	err := checkCompanyRole("wrong")
	require.ErrorContains(t, err, "invalid company role")
}
