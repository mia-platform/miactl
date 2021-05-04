package login

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewLoginCmd(t *testing.T) {
	const (
		username   = "random"
		password   = "secret"
		appID      = "film"
		providerID = "provia"
		baseURL    = "http://auth-provider/"
	)

	t.Run("successful login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off() // Flush pending mocks after test execution

		const expectedAccessToken = "YWNjZXNzVG9rZW4="

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(200).
			JSON(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		// Note: this is not testing the whole cli,
		// which means that interactions with global
		// flags must be tested in the main cmd package
		cmd := NewLoginCmd()
		cmd.Flags().Set("username", username)
		cmd.Flags().Set("password", password)
		cmd.Flags().Set("app-id", appID)
		cmd.Flags().Set("provider-id", providerID)

		require.Nil(t, cmd.Execute())

		accessToken := viper.GetString("apitoken")
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")

		require.True(t, gock.IsDone())
	})

	t.Run("failed login", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off() // Flush pending mocks after test execution

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(401).
			JSON(map[string]interface{}{})

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd := NewLoginCmd()
		cmd.Flags().Set("username", username)
		cmd.Flags().Set("password", password)
		cmd.Flags().Set("app-id", appID)
		cmd.Flags().Set("provider-id", providerID)

		err := cmd.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")

		accessToken := viper.GetString("apitoken")
		fmt.Println(accessToken)
		require.Empty(t, accessToken, "Access token must be empty string")

		require.True(t, gock.IsDone())
	})

	t.Run("failed login due to missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off() // Flush pending mocks after test execution

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(200).
			JSON(map[string]interface{}{
				"accessToken":  "YWNjZXNzVG9rZW4=",
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		cmd := NewLoginCmd()
		cmd.Flags().Set("username", username)
		cmd.Flags().Set("password", password)
		cmd.Flags().Set("app-id", appID)
		cmd.Flags().Set("provider-id", providerID)

		require.EqualError(t, cmd.Execute(), "API base URL not specified nor configured")

		accessToken := viper.GetString("apitoken")
		require.Empty(t, accessToken, "Access token differs from expected")

		require.False(t, gock.IsDone())
	})
}

func TestLogin(t *testing.T) {
	const (
		username   = "random"
		password   = "secret"
		appID      = "film"
		providerID = "provia"
		baseURL    = "http://auth-provider/"
	)

	t.Run("successful login", func(t *testing.T) {
		defer gock.Off() // Flush pending mocks after test execution

		const expectedAccessToken = "YWNjZXNzVG9rZW4="

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(200).
			JSON(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})

		accessToken, err := login(baseURL, username, password, appID, providerID)

		require.Nil(t, err)
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")

		require.True(t, gock.IsDone())
	})

	t.Run("failed login", func(t *testing.T) {
		defer gock.Off() // Flush pending mocks after test execution

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(401).
			JSON(map[string]string{})

		accessToken, err := login(baseURL, username, password, appID, providerID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")
		require.Empty(t, accessToken, "Access token must be empty string")

		require.True(t, gock.IsDone())
	})
}
