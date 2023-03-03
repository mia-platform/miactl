package login

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/skratchdot/open-golang/open"
)

type tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type Provider struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

const (
	appID       = "miactl"
	callbackUrl = "127.0.0.1:53535"
)

func GetTokensWithOIDC(endpoint string, providerID string) (*tokens, error) {

	jsonClient, err := jsonclient.New(jsonclient.Options{BaseURL: fmt.Sprintf("%s/api/", endpoint)})
	if err != nil {
		fmt.Printf("%v", "error generating JsonClient")
	}

	listener, err := net.Listen("tcp4", callbackUrl)
	if err != nil {
		return &tokens{}, err
	}

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", endpoint, q.Encode())

	if err := open.Run(startURL); err != nil {
		fmt.Println("Failed to open browser:", err)
		fmt.Println("Please open the following URL in your browser and complete the authentication process:")
		fmt.Println(startURL)
	}

	var code string
	var state string

	callbackPath := "/oauth/callback"
	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != callbackPath || req.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer listener.Close()

		qs := req.URL.Query()
		code = qs.Get("code")
		state = qs.Get("state")

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("You are successfully authenticated"))
	}))

	req, err := jsonClient.NewRequest(http.MethodPost, "oauth/token", map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return &tokens{}, err
	}

	token := &tokens{}
	_, err = jsonClient.Do(req, token)
	if err != nil {
		return &tokens{}, err
	}

	return token, nil
}
