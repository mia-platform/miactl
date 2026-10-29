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

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/clioptions"
)

func RemoveCmd(opts *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove CONTEXT [flags]",
		Short: "Remove a context",
		Long: `Remove a context. If the credential linked to the context is not used by any
other context it will be removed too, together with its cached token.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = opts.MiactlConfig
			return removeContext(args[0], locator)
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			locator := cliconfig.NewConfigPathLocator()
			config, err := locator.ReadConfig()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listContexts(config), cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func removeContext(name string, locator *cliconfig.ConfigPathLocator) error {
	config, err := locator.ReadConfig()
	if err != nil {
		return err
	}

	ctxConfig, found := config.Contexts[name]
	if !found {
		return fmt.Errorf("no context named \"%s\" exists", name)
	}
	authName := ctxConfig.AuthName

	delete(config.Contexts, name)

	wasCurrent := config.CurrentContext == name
	if wasCurrent {
		config.CurrentContext = ""
	}

	orphanedAuth := removeOrphanedAuth(config, ctxConfig, authName)

	if err := locator.WriteConfig(config); err != nil {
		return err
	}

	fmt.Printf("context \"%s\" successfully removed\n", name)
	if orphanedAuth {
		fmt.Printf("credential \"%s\" was no longer used by any context and was removed\n", authName)
	}
	if wasCurrent {
		fmt.Println("no context is currently active, use \"miactl context use\" to select one")
	}
	return nil
}

// removes auth configurations that are not used anymore
func removeOrphanedAuth(config *api.Config, ctxConfig *api.ContextConfig, authName string) bool {
	if authName == "" {
		return false
	}

	auth, found := config.Auth[authName]
	if !found || authStillReferenced(config, authName) {
		return false
	}

	cliconfig.RemoveCachedToken(ctxConfig, auth)
	delete(config.Auth, authName)
	return true
}

func authStillReferenced(config *api.Config, authName string) bool {
	for _, ctxConfig := range config.Contexts {
		if ctxConfig.AuthName == authName {
			return true
		}
	}
	return false
}
