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
	"fmt"

	"github.com/imdario/mergo"
	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func AuthCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth NAME",
		Short: "Set an auth configuration for miactl",
		Long: `Set an auth configuration for miactl. You can set service account access
and then attach it to one or more contexts.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
	options.AddBasicAuthFlags(flags)

	// add sub commands

	return cmd
}

func setAuth(authName string, opts *clioptions.CLIOptions) (bool, error) {
	locator := cliconfig.NewConfigPathLocator()
	locator.ExplicitPath = opts.MiactlConfig

	config, err := locator.ReadConfig()
	if err != nil {
		return false, err
	}

	newAuth := &api.AuthConfig{
		ClientID:     opts.BasicClientID,
		ClientSecret: opts.BasicClientSecret,
	}

	authConfig, found := config.Auth[authName]
	if !found {
		authConfig = new(api.AuthConfig)
	}

	if err := mergo.Merge(authConfig, newAuth, mergo.WithOverride); err != nil {
		return false, err
	}

	if config.Auth != nil {
		config.Auth[authName] = authConfig
	} else {
		config.Auth = map[string]*api.AuthConfig{
			authName: authConfig,
		}
	}

	return found, locator.WriteConfig(config)
}
