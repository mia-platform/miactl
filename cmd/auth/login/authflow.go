package login

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/davidebianchi/go-jsonclient"
)

type oauth struct {
	OpenBrowser func(url string) error
	HTTPClient  *jsonclient.Client

	localServerAddress string
}

type tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func (oauth oauth) localServerFlow(authHost, appID, providerID string) (*tokens, error) {
	listener, err := net.Listen("tcp4", oauth.localServerAddress)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", authHost, q.Encode())

	err = oauth.OpenBrowser(startURL)
	if err != nil {
		return nil, err
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

	req, err := oauth.HTTPClient.NewRequest(http.MethodPost, "oauth/token", map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return nil, err
	}

	token := &tokens{}
	_, err = oauth.HTTPClient.Do(req, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
