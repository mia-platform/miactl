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
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/configpath"
	"golang.org/x/oauth2"
)

type AuthReadWriter struct {
	locator *ConfigPathLocator
	config  *api.ContextConfig
}

func NewAuthReadWriter(locator *ConfigPathLocator, config *api.ContextConfig) *AuthReadWriter {
	return &AuthReadWriter{
		locator: locator,
		config:  config,
	}
}

func (rw *AuthReadWriter) ReadJWTToken() *oauth2.Token {
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(rw.config.Endpoint)))
	tokenPath := filepath.Join(configpath.CacheFolderPath(), cacheKey)
	tokenData, err := os.ReadFile(tokenPath)
	if err != nil {
		return &oauth2.Token{}
	}

	decoder := json.NewDecoder(bytes.NewBuffer(tokenData))
	jwt := new(oauth2.Token)
	if err := decoder.Decode(&jwt); err != nil {
		return &oauth2.Token{}
	}
	return jwt
}

func (rw *AuthReadWriter) WriteJWTToken(jwt *oauth2.Token) {
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(rw.config.Endpoint)))
	tokenPath := filepath.Join(configpath.CacheFolderPath(), cacheKey)
	dir := filepath.Dir(tokenPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	jwtBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(jwtBuffer)

	if err := encoder.Encode(jwt); err != nil {
		return
	}
	_ = os.WriteFile(tokenPath, jwtBuffer.Bytes(), 0600)
}
