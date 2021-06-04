package auth

import (
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
)

const miactlAppID = "miactl"

type AuthClient struct {
	JSONClient *jsonclient.Client
}

type IAuth interface {
	Login(string, string, string) (string, error)
}

type tokenRequest struct {
	GrantType  string `json:"grant_type"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	AppID      string `json:"appId"`
	ProviderID string `json:"providerId"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireAt     int64  `json:"expiresAt"`
}

func (a AuthClient) Login(username, password, providerID string) (string, error) {
	data := tokenRequest{
		GrantType:  "password",
		Username:   username,
		Password:   password,
		AppID:      miactlAppID,
		ProviderID: providerID,
	}

	loginReq, err := a.JSONClient.NewRequest(http.MethodPost, "/api/oauth/token", data)
	if err != nil {
		return "", fmt.Errorf("error creating login request: %w", err)
	}
	var loginResponse tokenResponse

	response, err := a.JSONClient.Do(loginReq, &loginResponse)
	if err != nil {
		return "", fmt.Errorf("auth error: %w", err)
	}
	defer response.Body.Close()

	return loginResponse.AccessToken, nil
}
