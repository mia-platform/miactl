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

package cliconfig

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/client"
)

type ConfigReader struct {
	config    *api.Config
	overrides *ConfigOverrides
}

func NewConfigReader(config *api.Config, overrides *ConfigOverrides) *ConfigReader {
	return &ConfigReader{
		config:    config,
		overrides: overrides,
	}
}

func (cr *ConfigReader) ClientConfig(locator *ConfigPathLocator) (*client.Config, error) {
	context, err := cr.getContext()
	if err != nil {
		return nil, err
	}

	authConfig, found := cr.getAuthConfig()

	clientConfig := &client.Config{
		Host: context.Endpoint,
		TLSClientConfig: client.TLSClientConfig{
			CAFile:   context.CertificateAuthority,
			Insecure: context.InsecureSkipTLSVerify,
		},
		AuthCacheReadWriter: NewAuthReadWriter(locator, context, authConfig),
		CompanyID:           context.CompanyID,
		ProjectID:           context.ProjectID,
	}

	if found {
		clientConfig.AuthConfig = client.AuthConfig{
			ClientID:     authConfig.ClientID,
			ClientSecret: authConfig.ClientSecret,
		}
	}
	return clientConfig, nil
}

func (cr *ConfigReader) getContext() (*api.ContextConfig, error) {
	currentContext, required := cr.getCurrentContextName()
	mergedContext := new(api.ContextConfig)
	if context, found := cr.config.Contexts[currentContext]; found {
		_ = mergo.Merge(mergedContext, context, mergo.WithOverride)
	} else if required {
		return nil, fmt.Errorf("context %s not found", currentContext)
	}

	if cr.overrides != nil {
		overrides := &api.ContextConfig{
			Endpoint:              cr.overrides.Endpoint,
			CertificateAuthority:  cr.overrides.CertificateAuthority,
			InsecureSkipTLSVerify: cr.overrides.InsecureSkipTLSVerify,
			CompanyID:             cr.overrides.CompanyID,
			ProjectID:             cr.overrides.ProjectID,
		}
		_ = mergo.Merge(mergedContext, overrides, mergo.WithOverride)
	}

	return mergedContext, nil
}

func (cr *ConfigReader) getCurrentContextName() (string, bool) {
	if cr.overrides != nil && len(cr.overrides.Context) > 0 {
		return cr.overrides.Context, true
	}

	return cr.config.CurrentContext, false
}

func (cr *ConfigReader) getAuthConfig() (*api.AuthConfig, bool) {
	authConfigName := cr.getAuthConfigName()
	if len(authConfigName) == 0 {
		return new(api.AuthConfig), true
	}

	authConfig, found := cr.config.Auth[authConfigName]
	return authConfig, found
}

func (cr *ConfigReader) getAuthConfigName() string {
	context, _ := cr.getContext()
	return context.AuthName
}
