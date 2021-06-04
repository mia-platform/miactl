package login

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/mia-platform/miactl/cmd/internal"
	"github.com/mia-platform/miactl/factory"
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
		expectedAccessToken = "YWNjZXNzVG9rZW4tMg=="
	)

	t.Run("successful login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, "/api/oauth/token", handler, nil)
		defer server.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(username, password, providerID)
		err := cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
		require.Equal(t, 1, callsCount)
	})

	t.Run("failed login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, "/api/oauth/token", handler, nil)
		defer server.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(username, password, providerID)
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")

		accessToken := viper.GetString("apitoken")
		fmt.Println(accessToken)
		require.Empty(t, accessToken, "Access token must be empty string")

		require.Equal(t, 1, callsCount)
	})

	t.Run("failed login due to missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, "/api/oauth/token", handler, nil)
		defer server.Close()

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		cmd, ctx := getLoginCommand(username, password, providerID)
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token differs from expected")

		require.Equal(t, 0, callsCount)
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
