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

package get

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/mia-platform/miactl/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	config1 = `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	config2 = `contexts:
  fake-ctx:
    apibaseurl: http://url
    projectid: "123"`
)

func TestGetProjects(t *testing.T) {
	opts := &clioptions.CLIOptions{}
	viper.SetConfigType("yaml")

	server := testutils.CreateMockServer()
	server.Start()
	defer server.Close()

	type TestCase struct {
		name        string
		config      string
		miaClient   *httphandler.MiaClient
		expectedErr string
	}
	testCases := []TestCase{
		{
			name:      "valid config, successful get",
			config:    config1,
			miaClient: httphandler.FakeMiaClient(fmt.Sprintf("%s/getprojects", server.URL)),
		},
		{
			name:        "invalid response body",
			config:      config1,
			miaClient:   httphandler.FakeMiaClient(fmt.Sprintf("%s/invalidbody", server.URL)),
			expectedErr: "invalid character",
		},
		{
			name:        "status code != 200",
			config:      config1,
			miaClient:   httphandler.FakeMiaClient(fmt.Sprintf("%s/notfound", server.URL)),
			expectedErr: "404 Not Found",
		},
		{
			name:        "company ID unset",
			config:      config2,
			miaClient:   httphandler.FakeMiaClient(fmt.Sprintf("%s/getprojects", server.URL)),
			expectedErr: "please set a company ID",
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := viper.ReadConfig(strings.NewReader(tc.config))
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}
		err = getProjects(tc.miaClient, opts)
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
	}
}
