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

package api

type Config struct {
	Contexts       map[string]*ContextConfig `yaml:"contexts"`
	CurrentContext string                    `yaml:"current-context"` //nolint:tagliatelle
	Auth           map[string]*AuthConfig    `yaml:"credentials"`     //nolint:tagliatelle
}

type ContextConfig struct {
	Endpoint              string `yaml:"endpoint"`
	CertificateAuthority  string `yaml:"certificate-authority,omitempty"`    //nolint:tagliatelle
	InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"` //nolint:tagliatelle
	CompanyID             string `yaml:"company-id,omitempty"`               //nolint:tagliatelle
	ProjectID             string `yaml:"project-id,omitempty"`               //nolint:tagliatelle
	AuthName              string `yaml:"credential,omitempty"`               //nolint:tagliatelle
	Environment           string `yaml:"environment,omitempty"`
}

type AuthConfig struct {
	ClientID          string `yaml:"client-id,omitempty"`        //nolint:tagliatelle
	ClientSecret      string `yaml:"client-secret,omitempty"`    //nolint:tagliatelle
	JWTKeyID          string `yaml:"key-id,omitempty"`           //nolint:tagliatelle
	JWTPrivateKeyData string `yaml:"private-key-data,omitempty"` //nolint:tagliatelle
}

func NewConfig() *Config {
	return &Config{
		Contexts: make(map[string]*ContextConfig),
		Auth:     make(map[string]*AuthConfig),
	}
}
