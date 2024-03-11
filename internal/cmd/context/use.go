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

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func UseCmd(opts *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use CONTEXT [flags]",
		Short: "Select a context to use",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = opts.MiactlConfig
			newContext := args[0]
			if err := setCurrentContext(args[0], locator); err != nil {
				return err
			}

			fmt.Printf("Switched to context \"%s\"\n", newContext)
			return nil
		},
	}

	return cmd
}

func setCurrentContext(newContext string, locator *cliconfig.ConfigPathLocator) error {
	config, err := locator.ReadConfig()
	if err != nil {
		return err
	}

	if _, found := config.Contexts[newContext]; !found {
		return fmt.Errorf("no context named \"%s\" exists", newContext)
	}

	config.CurrentContext = newContext
	return locator.WriteConfig(config)
}
