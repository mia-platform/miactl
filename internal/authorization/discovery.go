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

package authorization

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/mia-platform/miactl/internal/client"
)

const (
	protectedResourceMetadataPath = "/.well-known/oauth-protected-resource/api"
)

// protectedResourceMetadata holds the relevant fields from RFC 9728.
type protectedResourceMetadata struct {
	AuthorizationServers []string `json:"authorization_servers"`
}

// discoverOAuthConfig fetches /.well-known/oauth-protected-resource from the API
// base URL (RFC 9728), extracts the first authorization server, and performs OIDC
// discovery on it (RFC 8414 / OpenID Connect Discovery). Returns a ready
// *oauth2.Config, or an error if the resource metadata is unavailable or OIDC
// discovery fails.
func discoverOAuthConfig(ctx context.Context, apiClient client.Interface) (*oauth2.Config, error) {
	response, err := apiClient.Get().APIPath(protectedResourceMetadataPath).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching resource metadata: %w", err)
	}
	if err := response.Error(); err != nil {
		return nil, fmt.Errorf("resource metadata not available: %w", err)
	}

	var metadata protectedResourceMetadata
	if err := response.ParseResponse(&metadata); err != nil {
		return nil, fmt.Errorf("parsing resource metadata: %w", err)
	}

	if len(metadata.AuthorizationServers) == 0 {
		return nil, fmt.Errorf("no authorization_servers listed in resource metadata")
	}

	oidcCtx := oidc.ClientContext(ctx, apiClient.HTTPClient())
	provider, err := oidc.NewProvider(oidcCtx, metadata.AuthorizationServers[0])
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery for %q: %w", metadata.AuthorizationServers[0], err)
	}

	return &oauth2.Config{
		ClientID: appID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{oidc.ScopeOpenID},
	}, nil
}

// getTokenWithOIDC runs the OAuth2 authorization code flow with PKCE using the
// endpoints from the provided oauth2.Config. It starts a local callback server,
// opens the browser via readyFn, waits for the authorization code, and exchanges
// it for a token.
func getTokenWithOIDC(ctx context.Context, oauthCfg *oauth2.Config, apiClient client.Interface, readyFn LocalServerReadyHandler) (*oauth2.Token, error) {
	listener, err := newListener([]string{"127.0.0.1:53535", "127.0.0.1:13535"})
	if err != nil {
		return nil, err
	}

	cfg := *oauthCfg
	cfg.RedirectURL = "http://" + listener.Addr().String() + callbackEndpointString

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()
	authURL := cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	authResp, err := startLocalServerForToken(ctx, authURL, listener, readyFn)
	if err != nil {
		return nil, err
	}

	if authResp.State != state {
		return nil, fmt.Errorf("state mismatch in OAuth2 callback")
	}

	exchangeCtx := context.WithValue(ctx, oauth2.HTTPClient, apiClient.HTTPClient()) //nolint:staticcheck
	return cfg.Exchange(exchangeCtx, authResp.Code, oauth2.VerifierOption(verifier))
}
