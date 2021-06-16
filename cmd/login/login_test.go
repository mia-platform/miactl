package login

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/internal/mocks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewLoginCmd(t *testing.T) {
	const (
		username            = "random"
		password            = "secret"
		appID               = "film"
		providerID          = "provia"
		baseURL             = "http://auth-provider/"
		endpoint            = "/api/oauth/token"
		expectedAccessToken = "YWNjZXNzVG9rZW4tMg=="
		serverCertPath      = "../../testdata/server-cert.pem"
		serverKeyPath       = "../../testdata/server-key.pem"
		caCertPath          = "../../testdata/ca-cert.pem"
	)

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

		cmd, ctx := getLoginCommand(username, password, providerID)
		err = cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
	})

	t.Run("successful login - insecure access", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		serverCfg := mocks.CertificatesConfig{
			CertPath: "../../testdata/server-cert.pem",
			KeyPath:  "../../testdata/server-key.pem",
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

		cmd, ctx := getLoginCommand(username, password, providerID)
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

		cmd, ctx := getLoginCommand(username, password, providerID)

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

		cmd, ctx := getLoginCommand(username, password, providerID)

		err = cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")
		require.Contains(t, err.Error(), "x509: certificate signed by unknown authority")

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

		cmd, ctx := getLoginCommand(username, password, providerID)
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

		cmd, ctx := getLoginCommand(username, password, providerID)
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token differs from expected")
	})
}

func getLoginCommand(username, password, providerID string) (*cobra.Command, context.Context) {
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
