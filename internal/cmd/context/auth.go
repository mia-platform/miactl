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

package context

import (
	"encoding/json"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

func AuthCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth NAME",
		Short: "Set an auth configuration for miactl",
		Long: `Set an auth configuration for miactl. You can set service account access
and then attach it to one or more contexts.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			authName := args[0]
			modified, err := setAuth(args[0], options)
			if err != nil {
				return err
			}

			if modified {
				fmt.Printf("Auth \"%s\" modified.\n", authName)
			} else {
				fmt.Printf("Auth \"%s\" created.\n", authName)
			}

			return nil
		},
	}

	// add cmd flags
	flags := cmd.Flags()
	options.AddContextAuthFlags(flags)

	// add sub commands

	return cmd
}

func setAuth(authName string, opts *clioptions.CLIOptions) (bool, error) {
	if len(opts.JWTJsonPath) > 0 && (len(opts.BasicClientID) > 0 || len(opts.BasicClientSecret) > 0) {
		return false, fmt.Errorf("is not possible to set both jwt and basic service account configs")
	}

	locator := cliconfig.NewConfigPathLocator()
	locator.ExplicitPath = opts.MiactlConfig
	if len(opts.JWTJsonPath) > 0 {
		return saveJWTServiceAccount(authName, opts.JWTJsonPath, locator)
	}
	return saveBasicServiceAccount(authName, opts.BasicClientID, opts.BasicClientSecret, locator)
}

func saveBasicServiceAccount(name, clientID, clientSecret string, locator *cliconfig.ConfigPathLocator) (bool, error) {
	config, err := locator.ReadConfig()
	if err != nil {
		return false, err
	}

	newAuth := &api.AuthConfig{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		JWTKeyID:          "",
		JWTPrivateKeyData: "",
	}

	authConfig, found := config.Auth[name]
	if !found {
		authConfig = new(api.AuthConfig)
	}

	if err := mergo.Merge(authConfig, newAuth, mergo.WithOverride); err != nil {
		return false, err
	}

	if config.Auth != nil {
		config.Auth[name] = authConfig
	} else {
		config.Auth = map[string]*api.AuthConfig{
			name: authConfig,
		}
	}

	return found, locator.WriteConfig(config)
}

func saveJWTServiceAccount(name, jwtJSONPath string, locator *cliconfig.ConfigPathLocator) (bool, error) {
	fileData, err := os.ReadFile(jwtJSONPath)
	if err != nil {
		return false, err
	}

	jwtServiceAccount := new(resources.JWTServiceAccountJSON)
	err = json.Unmarshal(fileData, jwtServiceAccount)
	if err != nil {
		return false, err
	}

	config, err := locator.ReadConfig()
	if err != nil {
		return false, err
	}

	authConfig, found := config.Auth[name]
	if !found {
		authConfig = new(api.AuthConfig)
	}

	newAuth := &api.AuthConfig{
		ClientID:          jwtServiceAccount.ClientID,
		ClientSecret:      "",
		JWTKeyID:          jwtServiceAccount.KeyID,
		JWTPrivateKeyData: jwtServiceAccount.PrivateKeyData,
	}

	if err := mergo.Merge(authConfig, newAuth, mergo.WithOverride); err != nil {
		return false, err
	}

	if config.Auth != nil {
		config.Auth[name] = authConfig
	} else {
		config.Auth = map[string]*api.AuthConfig{
			name: authConfig,
		}
	}

	return found, locator.WriteConfig(config)
}
