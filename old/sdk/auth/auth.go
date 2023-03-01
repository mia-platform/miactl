// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
)

const miactlAppID = "miactl"

type Client struct {
	JSONClient *jsonclient.Client
}

type IAuth interface {
	Login(string, string, string) (string, error)
}

type tokenRequest struct {
	GrantType  string `json:"grant_type"` //nolint:tagliatelle
	Username   string `json:"username"`
	Password   string `json:"password"`
	AppID      string `json:"appId"`
	ProviderID string `json:"providerId"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func (a Client) Login(username, password, providerID string) (string, error) {
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
