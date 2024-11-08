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

package rest

import (
	"net/http"

	"github.com/mia-platform/miactl/client/plugin/pkg/client/auth/serviceaccount"
	"github.com/mia-platform/miactl/client/transport"
	"golang.org/x/oauth2"
)

// HTTPClientForConfig return a new http.Client with the transport security provided in the config
// Will return the default http.DefaultClient if no special case behavior is needed.
func HTTPClientForConfig(config *Config) (*http.Client, error) {
	httpClient := http.DefaultClient

	transport, err := TransportForConfig(config)
	if err != nil {
		return nil, err
	}

	if transport != http.DefaultTransport || config.Timeout > 0 {
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		}
	}

	return httpClient, nil
}

type noopProvider struct{}

func (*noopProvider) ReadJWTToken() *oauth2.Token { return nil }
func (*noopProvider) WriteJWTToken(*oauth2.Token) {}

// TransportForConfig return a new transport for the config or the one attached to it
func TransportForConfig(config *Config) (http.RoundTripper, error) {
	if config.Transport != nil {
		return config.Transport, nil
	}

	transportConfig := &transport.Config{
		UserAgent: config.UserAgent,
		TLSConfig: transport.TLSConfig{
			Insecure: config.Insecure,
			CAFile:   config.CAFile,
		},
		DisableCompression: config.DisableCompression,
		EnableDebug:        config.EnableDebug,
	}

	if config.AuthConfig != nil {
		authenticator, err := serviceaccount.NewServiceAccountAuthenticator(transportConfig, serviceaccount.AuthConfig{
			Server:         config.Host,
			ClientID:       config.AuthConfig.ClientID,
			ClientSecret:   config.AuthConfig.ClientSecret,
			KeyID:          config.AuthConfig.JWTKeyID,
			PrivateKeyData: config.AuthConfig.JWTPrivateKeyData,
		})
		if err != nil {
			return nil, err
		}

		transportConfig.Wrap(authenticator.Wrap)
	}

	if authProvider != nil {
		var cacheProvider AuthCacheReadWriter = &noopProvider{}
		if config.AuthCacheReadWriter != nil {
			cacheProvider = config.AuthCacheReadWriter
		}
		provider := authProvider(config, cacheProvider)
		transportConfig.Wrap(provider.Wrap)
	}

	return transport.New(transportConfig)
}
