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

package company

import (
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/mia-platform/miactl/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestNewGetCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewListCompaniesCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestListCompanies(t *testing.T) {
	server := testutils.CreateMockServer()
	server.Start()
	defer server.Close()

	type TestCase struct {
		name        string
		miaClient   *httphandler.MiaClient
		expectedErr string
	}
	testCases := []TestCase{
		{
			name:      "valid config, successful get",
			miaClient: httphandler.FakeMiaClient(fmt.Sprintf("%s/getcompanies", server.URL)),
		},
		{
			name:        "invalid response body",
			miaClient:   httphandler.FakeMiaClient(fmt.Sprintf("%s/invalidbody", server.URL)),
			expectedErr: "invalid character",
		},
		{
			name:        "status code != 200",
			miaClient:   httphandler.FakeMiaClient(fmt.Sprintf("%s/notfound", server.URL)),
			expectedErr: "404 Not Found",
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := listCompanies(tc.miaClient)
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
	}
}
