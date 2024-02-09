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

	"dario.cat/mergo"
	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func SetCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set CONTEXT [flags]",
		Short: "Set a context for miactl",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			contextName := args[0]
			modified, err := setContext(args[0], options)
			if err != nil {
				return err
			}

			if modified {
				fmt.Printf("Context \"%s\" modified.\n", contextName)
			} else {
				fmt.Printf("Context \"%s\" created.\n", contextName)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	options.AddConnectionFlags(flags)
	options.AddCompanyFlags(flags)
	options.AddProjectFlags(flags)
	options.AddEnvironmentFlags(flags)
	options.AddAuthFlags(flags)

	return cmd
}

func setContext(contextName string, opts *clioptions.CLIOptions) (bool, error) {
	locator := cliconfig.NewConfigPathLocator()
	locator.ExplicitPath = opts.MiactlConfig

	config, err := locator.ReadConfig()
	if err != nil {
		return false, err
	}

	newConfigContext := &api.ContextConfig{
		Endpoint:              opts.Endpoint,
		CertificateAuthority:  opts.CAFile,
		CompanyID:             opts.CompanyID,
		ProjectID:             opts.ProjectID,
		InsecureSkipTLSVerify: opts.Insecure,
		AuthName:              opts.Auth,
		Environment:           opts.Environment,
	}

	contextConfig, found := config.Contexts[contextName]
	if !found {
		contextConfig = new(api.ContextConfig)
	}

	if err := mergo.Merge(contextConfig, newConfigContext, mergo.WithOverride); err != nil {
		return false, err
	}

	if config.Contexts != nil {
		config.Contexts[contextName] = contextConfig
	} else {
		config.Contexts = map[string]*api.ContextConfig{
			contextName: contextConfig,
		}
	}

	return found, locator.WriteConfig(config)
}
