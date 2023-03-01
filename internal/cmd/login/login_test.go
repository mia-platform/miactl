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

package login

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/mocks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	username            = "random"
	password            = "secret"
	appID               = "film"
	providerID          = "provia"
	baseURL             = "http://auth-provider/"
	endpoint            = "/api/oauth/token"
	expectedAccessToken = "YWNjZXNzVG9rZW4tMg=="
	serverCertPath      = "../../../testdata/server-cert.pem"
	serverKeyPath       = "../../../testdata/server-key.pem"
	caCertPath          = "../../../testdata/ca-cert.pem"
)

func TestNewLoginCmd(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint:    endpoint,
				Method:      http.MethodPost,
				RequestBody: nil,
				Reply: map[string]interface{}{
					"accessToken":  expectedAccessToken,
					"refreshToken": "cmVmcmVzaFRva2Vu",
					"expiresAt":    1619799800,
				},
				ReplyStatus: http.StatusOK,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, nil)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)
		err = cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
	})

	t.Run("successful login - insecure access", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		serverCfg := mocks.CertificatesConfig{
			CertPath: "../../../testdata/server-cert.pem",
			KeyPath:  "../../../testdata/server-key.pem",
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint:    endpoint,
				Method:      http.MethodPost,
				RequestBody: nil,
				Reply: map[string]interface{}{
					"accessToken":  expectedAccessToken,
					"refreshToken": "cmVmcmVzaFRva2Vu",
					"expiresAt":    1619799800,
				},
				ReplyStatus: http.StatusOK,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)
		cmd.Flags().Set("insecure", "true")

		err = cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
	})

	t.Run("successful login - select custom CA certificate", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		serverCfg := mocks.CertificatesConfig{
			CertPath: serverCertPath,
			KeyPath:  serverKeyPath,
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint:    endpoint,
				Method:      http.MethodPost,
				RequestBody: nil,
				Reply: map[string]interface{}{
					"accessToken":  expectedAccessToken,
					"refreshToken": "cmVmcmVzaFRva2Vu",
					"expiresAt":    1619799800,
				},
				ReplyStatus: http.StatusOK,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("ca-cert", caCertPath)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)

		err = cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
	})

	t.Run("failed login - certificate issues", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		serverCfg := mocks.CertificatesConfig{
			CertPath: serverCertPath,
			KeyPath:  serverKeyPath,
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint:    endpoint,
				Method:      http.MethodPost,
				RequestBody: nil,
				Reply: map[string]interface{}{
					"accessToken":  expectedAccessToken,
					"refreshToken": "cmVmcmVzaFRva2Vu",
					"expiresAt":    1619799800,
				},
				ReplyStatus: http.StatusOK,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)

		err = cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")
		require.Regexp(t, regexp.MustCompile("x509: certificate signed by unknown authority|certificate is not standards compliant"), err)

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token must be empty string")
	})

	t.Run("failed login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint:    endpoint,
				Method:      http.MethodPost,
				RequestBody: nil,
				Reply:       map[string]interface{}{},
				ReplyStatus: http.StatusUnauthorized,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, nil)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)
		err = cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token must be empty string")
	})

	t.Run("failed login due to missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(t)
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token differs from expected")
	})
}

func getLoginCommand(t *testing.T) (*cobra.Command, context.Context) {
	t.Helper()
	// Note: this is not testing the whole cli,
	// which means that interactions with global
	// flags must be tested in the main cmd package
	cmd := NewLoginCmd()
	cmd.Flags().Set("username", username)
	cmd.Flags().Set("password", password)
	cmd.Flags().Set("provider-id", providerID)

	ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())

	return cmd, ctx
}
