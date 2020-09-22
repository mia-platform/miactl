package login

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"

	"github.com/davidebianchi/go-jsonclient"
)

type Oauth struct {
	OpenBrowser func(goos string, url string) error
	HTTPClient  *jsonclient.Client
}

// TODO: providerID should be selected from the console list of providers. It supports multiple providers.
func (oauth Oauth) localServerFlow(authHost, appID, providerID string) (string, error) {
	// TODO: switch to 0 to automatically select a port
	listener, err := net.Listen("tcp4", "127.0.0.1:53535")
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", authHost, q.Encode())

	err = oauth.OpenBrowser(runtime.GOOS, startURL)
	if err != nil {
		return "", err
	}

	var code string
	var state string

	callbackPath := "/oauth/callback"
	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != callbackPath {
			w.WriteHeader(404)
			return
		}
		defer listener.Close()

		qs := r.URL.Query()
		code = qs.Get("code")
		state = qs.Get("state")
		w.Header().Add("content-type", "text/html")

		w.WriteHeader(200)
		w.Write([]byte("You are successfully authenticated"))
	}))

	req, err := oauth.HTTPClient.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/oauth/token", authHost), map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return "", err
	}

	type OauthToken struct {
		AccessToken string `json:"accessToken"`
	}

	token := &OauthToken{}

	_, err = oauth.HTTPClient.Do(req, token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
