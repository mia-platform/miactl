package login

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLocalLoginOIDC(t *testing.T) {
	providerID := "the-provider"
	code := "my-code"
	state := "my-state"
	callbackUrl := "http://127.0.0.1:53535"
	endpoint := "http://127.0.0.1:53534"

	t.Run("correctly returns token", func(t *testing.T) {
		http.HandleFunc("/api/oauth/token", handleCallbackToken)
		go func() {
			err := http.ListenAndServe(":53534", nil)
			if err != nil {
				fmt.Println(err)
			}
		}()

		browser := fakeBrowser{
			code:        code,
			state:       state,
			callbackUrl: callbackUrl,
			endpoint:    endpoint,
		}
		// browserq := browser{}
		tokens, err := GetTokensWithOIDC(endpoint, providerID, browser)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(tokens)
	})
}

func handleCallbackToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{\"AccessToken\":\"accesstoken\", \"RefreshToken\":\"refreshToken\", \"ExpiresAt\":23345}"))
}
