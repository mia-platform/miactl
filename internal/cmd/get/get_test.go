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

func TestGetProjects(t *testing.T) {
	opts := &clioptions.CLIOptions{}
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

	server := testutils.CreateMockServer()
	server.Start()
	defer server.Close()

	mc1 := httphandler.FakeMiaClient(server.URL)

	err = getProjects(mc1, opts)
	require.NoError(t, err)

	// invalid response body
	mc2 := httphandler.FakeMiaClient(fmt.Sprintf("%s/invalidbody", server.URL))

	err = getProjects(mc2, opts)
	require.Error(t, err)

	// status code != 200
	mc3 := httphandler.FakeMiaClient(fmt.Sprintf("%s/notfound", server.URL))

	err = getProjects(mc3, opts)
	require.Error(t, err)

	// company ID unset
	config2 := `contexts:
  fake-ctx:
    apibaseurl: http://url
    projectid: "123"`
	err = viper.ReadConfig(strings.NewReader(config2))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	mc4 := httphandler.FakeMiaClient(server.URL)

	err = getProjects(mc4, opts)
	require.Error(t, err)
}
