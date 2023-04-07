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

package httphandler

import (
	"strings"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestClientBuilding(t *testing.T) {
	mExpected := MiaClient{
		SessionHandler: SessionHandler{
			url: "url",
		},
	}

	r := SessionHandler{
		url: "url",
	}

	m := NewMiaClientBuilder().
		WithSessionHandler(r)

	require.Equal(t, *m, mExpected)
}

func TestGetSession(t *testing.T) {
	sh := SessionHandler{
		url: "url",
	}
	mc := MiaClient{
		SessionHandler: sh,
	}
	actualSH := mc.GetSession()
	require.Equal(t, mc.SessionHandler, *actualSH)
}

func TestConfigureDefaultMiaClient(t *testing.T) {
	opts := &clioptions.CLIOptions{
		Endpoint:  "http://url",
		CompanyID: "123",
		ProjectID: "123",
	}
	viper.SetConfigType("yaml")
	config := `contexts:
  test-context:
    endpoint: http://url
    companyid: "123"
    projectid: "123"
current-context: test-context`
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}

	// valid mia client
	session := SessionHandler{
		url:     "http://url/test",
		context: testContext,
		client:  defaultClient,
		auth: &Auth{
			url:        "http://url",
			providerID: oktaProvider,
			browser:    login.Browser{},
		},
	}
	expectedMiaClient := &MiaClient{
		SessionHandler: session,
	}
	mc, err := ConfigureDefaultMiaClient(opts, testURI)
	require.NoError(t, err)
	require.EqualValues(t, expectedMiaClient, mc)

	// current context unset
	config = `contexts:
  test-context:
    endpoint: http://url
    companyid: "123"
    projectid: "123"`
	err = viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	mc, err = ConfigureDefaultMiaClient(opts, testURI)
	require.Nil(t, mc)
	require.Error(t, err)
}
