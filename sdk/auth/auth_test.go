package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	const (
		username            = "random"
		password            = "secret"
		appID               = "film"
		providerID          = "provia"
		baseURL             = "http://auth-provider/"
		expectedAccessToken = "YWNjZXNzVG9rZW4="
	)

	t.Run("successful login", func(t *testing.T) {
		called := false

		handler := func(w http.ResponseWriter, r *http.Request) {
			called = true
			data, _ := json.Marshal(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := getTestServer(t, "/api/oauth/token", handler, nil)
		defer server.Close()

		secureClient := getTestClient(t, server.URL, false)
		accessToken, err := secureClient.Login(username, password, providerID)

		require.Nil(t, err)
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
		require.True(t, called)
	})

	t.Run("successful login - insecure connection enabled", func(t *testing.T) {
		serverCfg := map[string]string{
			"cert": "../testdata/server-cert.pem",
			"key":  "../testdata/server-key.pem",
		}
		called := false

		handler := func(w http.ResponseWriter, r *http.Request) {
			called = true
			data, _ := json.Marshal(map[string]interface{}{
				"accessToken":  expectedAccessToken,
				"refreshToken": "cmVmcmVzaFRva2Vu",
				"expiresAt":    1619799800,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := getTestServer(t, "/api/oauth/token", handler, serverCfg)
		defer server.Close()

		insecureClient := getTestClient(t, server.URL, true)
		accessToken, err := insecureClient.Login(username, password, providerID)

		require.Nil(t, err)
		require.Equal(t, expectedAccessToken, accessToken, "Access token differs from expected")
		require.True(t, called)
	})

	t.Run("failed login", func(t *testing.T) {
		called := false

		handler := func(w http.ResponseWriter, r *http.Request) {
			called = true
			data, _ := json.Marshal(map[string]interface{}{})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(data)
		}
		server, _ := getTestServer(t, "/api/oauth/token", handler, nil)
		defer server.Close()

		secureClient := getTestClient(t, server.URL, false)
		accessToken, err := secureClient.Login(username, password, providerID)

		require.Error(t, err)
		require.Contains(t, err.Error(), "auth error:")
		require.Empty(t, accessToken, "Access token must be empty string")
		require.True(t, called)
	})
}

func getTestServer(t testing.TB, path string, h http.HandlerFunc, serverCfg map[string]string) (*httptest.Server, error) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(path, h)
	server := httptest.NewUnstartedServer(mux)

	if serverCfg != nil {
		cert, err := tls.LoadX509KeyPair(serverCfg["cert"], serverCfg["key"])
		if err != nil {
			return nil, err
		}
		server.TLS = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server.StartTLS()
	} else {
		server.Start()
	}

	return server, nil
}

func getTestClient(t *testing.T, url string, skipCertificate bool) IAuth {
	t.Helper()

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: skipCertificate,
	}
	if !strings.HasSuffix(url, "/") {
		url = fmt.Sprintf("%s/", url)
	}
	clientOptions := jsonclient.Options{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout:   time.Second * 10,
			Transport: customTransport,
		},
	}

	client, err := jsonclient.New(clientOptions)
	require.NoError(t, err, "error creating client")

	return AuthClient{
		JSONClient: client,
	}
}
