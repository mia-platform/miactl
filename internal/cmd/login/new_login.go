package login

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/davidebianchi/go-jsonclient"
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

var (
	state string
	code  string
)

func GetTokensWithOIDC(endpoint string, providerID string, b browserI) (*tokens, error) {
	jsonClient, err := jsonclient.New(jsonclient.Options{BaseURL: fmt.Sprintf("%s/api/", endpoint)})
	if err != nil {
		fmt.Printf("%v", "error generating JsonClient")
	}
	callbackPath := "/oauth/callback"

	// http.HandleFunc(callbackPath, handleCallback)

	// go func() {
	// 	err = http.ListenAndServe(":53535", nil)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }()

	l, err := net.Listen("tcp", ":53535")
	if err != nil {
		panic(err)
	}
	// Server HTTP request
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == callbackPath && r.Method == http.MethodGet:
				handleCallback(w, r)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	}
	defer s.Close()
	go s.Serve(l) // actually start the server, it manges all buffered requests

	q := url.Values{}
	q.Set("appId", appID)
	q.Set("providerId", providerID)

	startURL := fmt.Sprintf("%s/api/authorize?%s", endpoint, q.Encode())

	b.open(startURL)

	req, err := jsonClient.NewRequest(http.MethodPost, "oauth/token", map[string]interface{}{
		"code":  code,
		"state": state,
	})
	if err != nil {
		return &tokens{}, err
	}

	fmt.Println(jsonClient)

	token := &tokens{}
	_, err = jsonClient.Do(req, token)
	if err != nil {
		return &tokens{}, err
	}

	fmt.Println("token", token)
	fmt.Println("error", err)

	fmt.Println(err)
	return token, nil
}

func handleCallback(w http.ResponseWriter, req *http.Request) {
	// if req.URL.Path != callbackPath || req.Method != http.MethodGet {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	qs := req.URL.Query()
	code = qs.Get("code")
	state = qs.Get("state")

	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusBadGateway)
}
