package sdk

import (
	"fmt"
	"testing"

	sdkErrors "github.com/mia-platform/miactl/sdk/errors"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("throws with empty options", func(t *testing.T) {
		tests := []struct {
			option Options
		}{
			{option: Options{}},
			{option: Options{APICookie: "cookie", APIKey: "sid=asd"}},
			{option: Options{APIToken: "token"}},
		}
		for _, test := range tests {
			client, err := New(test.option)
			require.EqualError(t, err, fmt.Sprintf("%s: client options are not correct", sdkErrors.ErrCreateClient))
			require.Nil(t, client)
		}
	})

	t.Run("throws with wrong base url", func(t *testing.T) {
		client, err := New(Options{
			APIBaseURL: "wrong	",
			APIKey:     "apiKey",
			APICookie:  "sid=asd",
		})
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("throws due to wrong certificate path", func(t *testing.T) {
		client, err := New(Options{
			APIBaseURL:            "http://my-url/path/",
			APIToken:              "api-token",
			SkipCertificate:       false,
			AdditionalCertificate: "./testdata/missing-ca-cert.pem",
		})
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("correctly returns mia client", func(t *testing.T) {
		tests := []struct {
			option Options
		}{
			{
				option: Options{
					APIBaseURL: "http://my-url/path/",
					APIKey:     "my apiKey",
					APICookie:  "sid=asd",
				},
			},
			{
				option: Options{
					APIBaseURL: "http://my-url/path/",
					APIToken:   "api-token",
				},
			},
			{
				option: Options{
					APIBaseURL:      "http://my-url/path/",
					APIToken:        "api-token",
					SkipCertificate: true,
				},
			},
			{
				option: Options{
					APIBaseURL:            "http://my-url/path/",
					APIToken:              "api-token",
					SkipCertificate:       false,
					AdditionalCertificate: "../testdata/ca-cert.pem",
				},
			},
		}
		for _, test := range tests {
			client, err := New(test.option)
			checkClient(t, client, err)
		}
	})
}

func checkClient(t testing.TB, client *MiaClient, err error) {
	require.NoError(t, err, "new client error")
	require.NotNil(t, client)
	require.NotNil(t, client.Auth)
	require.NotNil(t, client.Deploy)
	require.NotNil(t, client.Projects)
}
