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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	valid = `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"
current-context: fake-ctx`
	noCompanyID = `contexts:
  fake-ctx:
    apibaseurl: http://url
    projectid: "123"`
	noCurrCtx = `contexts:
  fake-ctx:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
)

type TestCase struct {
	name        string
	config      string
	arg         string
	expectedOut string
	expectedErr string
}

func TestNewContextCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewContextCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestGetContextBaseURL(t *testing.T) {
	viper.SetConfigType("yaml")

	testCases := []TestCase{
		{
			name:        "valid config and context",
			config:      valid,
			arg:         "fake-ctx",
			expectedOut: "http://url",
			expectedErr: "",
		},
		{
			name:        "wrong context",
			config:      valid,
			arg:         "wrong-ctx",
			expectedOut: "",
			expectedErr: "context wrong-ctx does not exist",
		},
	}

	for _, tc := range testCases {
		err := viper.ReadConfig(strings.NewReader(tc.config))
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}
		url, err := GetContextBaseURL(tc.arg)
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
		require.Equal(t, tc.expectedOut, url)
	}

}

func TestGetContextCompanyID(t *testing.T) {
	viper.SetConfigType("yaml")

	testCases := []TestCase{
		{
			name:        "valid context, existing company ID",
			config:      valid,
			arg:         "fake-ctx",
			expectedOut: "123",
			expectedErr: "",
		},
		{
			name:        "wrong context name",
			config:      valid,
			arg:         "wrong-ctx",
			expectedOut: "",
			expectedErr: "context wrong-ctx does not exist",
		},
		{
			name:        "company id unset",
			config:      noCompanyID,
			arg:         "fake-ctx",
			expectedOut: "",
			expectedErr: "please set a company ID",
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := viper.ReadConfig(strings.NewReader(tc.config))
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}
		companyID, err := GetContextCompanyID(tc.arg)
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
		require.Equal(t, tc.expectedOut, companyID)
	}

}

func TestGetCurrentContext(t *testing.T) {
	viper.SetConfigType("yaml")

	testCases := []TestCase{
		{
			name:        "current context set",
			config:      valid,
			expectedOut: "fake-ctx",
			expectedErr: "",
		},
		{
			name:        "current context unset",
			config:      noCurrCtx,
			expectedOut: "",
			expectedErr: "current context is unset",
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		err := viper.ReadConfig(strings.NewReader(tc.config))
		if err != nil {
			t.Fatalf("unexpected error reading config: %v", err)
		}
		currentCtx, err := GetCurrentContext()
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.expectedErr)
		}
		require.Equal(t, tc.expectedOut, currentCtx)
	}

}

func TestSetContextValues(t *testing.T) {
	viper.New()
	viper.Set("contexts.current.projectid", "projectid")
	viper.Set("contexts.current.companyid", "companyid")
	viper.Set("contexts.current.apibaseurl", "apibaseurl")
	viper.Set("contexts.current.cacert", "cacert")

	type fakeOptsValues struct {
		ProjectID  string
		CompanyID  string
		APIBaseURL string
		CACert     string
	}

	f := fakeOptsValues{}

	fakeCommand := &cobra.Command{
		Use: "fake",
	}

	fakeCommand.Flags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
	fakeCommand.Flags().StringVar(&f.APIBaseURL, "apibaseurl", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	fakeCommand.Flags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
	fakeCommand.Flags().StringVar(
		&f.CACert,
		"ca-cert",
		"",
		"file path to a CA certificate, which can be employed to verify server certificate",
	)

	t.Run("test keep values from config file", func(t *testing.T) {
		SetContextValues(fakeCommand, "current")

		require.Equal(t, fakeCommand.Flag("project-id").Value.String(), "projectid")
		require.Equal(t, fakeCommand.Flag("company-id").Value.String(), "companyid")
		require.Equal(t, fakeCommand.Flag("apibaseurl").Value.String(), "apibaseurl")
	})

	t.Run("test set values from clioptions", func(t *testing.T) {
		f.APIBaseURL = "newapibaseurl"
		f.CompanyID = "newcompanyid"
		f.ProjectID = "newprojectid"

		SetContextValues(fakeCommand, "current")

		require.Equal(t, fakeCommand.Flag("project-id").Value.String(), "newprojectid")
		require.Equal(t, fakeCommand.Flag("company-id").Value.String(), "newcompanyid")
		require.Equal(t, fakeCommand.Flag("apibaseurl").Value.String(), "newapibaseurl")
	})
}
