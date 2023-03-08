package login

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalLoginOIDC(t *testing.T) {
	providerID := "the-provider"
	code := "my-code"
	state := "my-state"
	endpoint := "http://127.0.0.1:53534"
	callbackPath := "/api/oauth/token"

	t.Run("correctly returns token", func(t *testing.T) {
		l, err := net.Listen("tcp", ":53534")
		if err != nil {
			panic(err)
		}

		s := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err != nil {
					fmt.Println(err)
				}
				var data map[string]interface{}
				err = json.NewDecoder(r.Body).Decode(&data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				switch {
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && data["code"] == code && data["state"] == state:
					handleCallbackSuccesfulToken(w, r)
					return
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && (data["code"] != code || data["state"] != state):
					handleCallbackUnsuccesfulToken(w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}),
		}
		defer s.Close()

		go s.Serve(l)
		expectedToken := tokens{
			AccessToken:  "accesstoken",
			RefreshToken: "refreshToken",
			ExpiresAt:    23345,
		}

		browser := fakeBrowser{
			code:        code,
			state:       state,
			callbackUrl: callbackUrl,
		}
		tokens, err := GetTokensWithOIDC(endpoint, providerID, browser)
		if err != nil {
			fmt.Println(err)
		}
		require.Equal(t, *tokens, expectedToken)

	})

	t.Run("return error with incorrect callback", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		l, err := net.Listen("tcp", ":53534")
		if err != nil {
			panic(err)
		}

		s := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err != nil {
					fmt.Println(err)
				}
				var data map[string]interface{}
				err = json.NewDecoder(r.Body).Decode(&data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				switch {
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && data["code"] == code && data["state"] == state:
					handleCallbackSuccesfulToken(w, r)
					return
				case r.URL.Path == callbackPath && r.Method == http.MethodPost && (data["code"] != code || data["state"] != state):
					handleCallbackUnsuccesfulToken(w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}),
		}
		defer s.Close()

		go s.Serve(l)
		callbackUrl := "http://127.0.0.1:45536"
		browser := fakeBrowser{
			code:        code,
			state:       state,
			callbackUrl: callbackUrl,
		}
		_, err = GetTokensWithOIDC(callbackUrl, providerID, browser)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(err)
		require.Error(t, err)
	})

}
func TestOpenBrowser(t *testing.T) {
	t.Run("return error with incorrect provider url", func(t *testing.T) {
		incorrectUrl := "incorrect"
		browser := browser{}
		err := browser.open(incorrectUrl)
		require.Error(t, err)

	})

}

func handleCallbackSuccesfulToken(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{\"AccessToken\":\"accesstoken\", \"RefreshToken\":\"refreshToken\", \"ExpiresAt\":23345}"))
}

func handleCallbackUnsuccesfulToken(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusForbidden)

}
