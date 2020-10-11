package login

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestLocalServerFlow(t *testing.T) {
	providerID := "the-provider"
	expectedAccessToken := "my-access-token"
	expectedRefreshToken := "my-refresh-token"
	expectedExpiresAt := time.Now().Unix()
	code := "my-code"
	state := "my-state"

	t.Run("returns error if OpenBrowser returns error", func(t *testing.T) {
		var expectedError = fmt.Errorf("my error")
		o := oauth{
			OpenBrowser: func(url string) error {
				return expectedError
			},
		}

		accessToken, err := o.localServerFlow("", appID, providerID)
		require.EqualError(t, err, expectedError.Error())
		require.Nil(t, accessToken)
	})

	t.Run("correctly returns token", func(t *testing.T) {
		localServerAddress := "127.0.0.1:44565"

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			t.Helper()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"accessToken":"%s","refreshToken":"%s","expiresAt":%d}`, expectedAccessToken, expectedRefreshToken, expectedExpiresAt)))
		}))

		jsonClient, err := jsonclient.New(jsonclient.Options{
			BaseURL: fmt.Sprintf("%s/api/", s.URL),
		})
		require.NoError(t, err)

		o := oauth{
			OpenBrowser: func(url string) error {
				require.Equal(t,
					fmt.Sprintf("%s/api/authorize?appId=%s&providerId=%s", s.URL, appID, providerID),
					url,
					"browser not opened correctly",
				)

				go func(t *testing.T) {
					resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/oauth/callback?code=%s&state=%s", localServerAddress, code, state))
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, resp.StatusCode)

					body, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)
					require.Equal(t, "You are successfully authenticated", string(body))
				}(t)
				return nil
			},
			HTTPClient: jsonClient,

			localServerAddress: localServerAddress,
		}

		token, err := o.localServerFlow(s.URL, appID, providerID)
		require.NoError(t, err)
		require.Equal(t, &tokens{
			AccessToken:  expectedAccessToken,
			RefreshToken: expectedRefreshToken,
			ExpiresAt:    expectedExpiresAt,
		}, token)
	})

	t.Run("throws if api calls fails", func(t *testing.T) {
		localServerAddress := "127.0.0.1:44566"

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			t.Helper()
			w.WriteHeader(http.StatusBadRequest)
			w.Write(nil)
		}))

		jsonClient, err := jsonclient.New(jsonclient.Options{
			BaseURL: fmt.Sprintf("%s/api/", s.URL),
		})
		require.NoError(t, err)

		o := oauth{
			OpenBrowser: func(url string) error {
				require.Equal(t,
					fmt.Sprintf("%s/api/authorize?appId=%s&providerId=%s", s.URL, appID, providerID),
					url,
					"browser not opened correctly",
				)

				go func(t *testing.T) {
					resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s/oauth/callback?code=%s&state=%s", localServerAddress, code, state))
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, resp.StatusCode)

					body, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)
					require.Equal(t, "You are successfully authenticated", string(body))
				}(t)
				return nil
			},
			HTTPClient: jsonClient,

			localServerAddress: localServerAddress,
		}

		token, err := o.localServerFlow(s.URL, appID, providerID)
		require.EqualError(t, err, fmt.Sprintf("POST %s/api/oauth/token: 400", s.URL))
		require.Nil(t, token)
	})

	t.Run("returns error if localServerAddress is not correct", func(t *testing.T) {
		o := oauth{
			localServerAddress: "not-correct",
		}
		tokens, err := o.localServerFlow("", appID, providerID)

		require.Nil(t, tokens)
		require.EqualError(t, err, "listen tcp4: address not-correct: missing port in address")
	})
}
