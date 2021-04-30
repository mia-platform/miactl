package login

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

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
		require.Equal(t, expectedAccessToken, accessToken)

		require.True(t, gock.IsDone())
	})

	t.Run("failed login", func(t *testing.T) {
		defer gock.Off() // Flush pending mocks after test execution

		gock.New(baseURL).
			Post("/oauth/token").
			Reply(401).
			JSON(map[string]string{})

		accessToken, err := login(baseURL, username, password, appID, providerID)

		require.Error(t, err, "auth error:")
		require.Equal(t, "", accessToken)

		require.True(t, gock.IsDone())
	})
}
