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
	"io"
	"sort"

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func ListCmd(opts *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List available contexts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			locator := cliconfig.NewConfigPathLocator()
			locator.ExplicitPath = opts.MiactlConfig

			return printContexts(cmd.OutOrStdout(), locator)
		},
	}

	return cmd
}

func listContexts(config *api.Config) []string {
	contextNames := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	return contextNames
}

func printContexts(out io.Writer, locator *cliconfig.ConfigPathLocator) error {
	config, err := locator.ReadConfig()
	if err != nil {
		return err
	}

	contextNames := listContexts(config)
	currentContext := config.CurrentContext
	for _, key := range contextNames {
		switch key {
		case currentContext:
			fmt.Fprintf(out, "* %s\n", key)
		default:
			fmt.Fprintf(out, "  %s\n", key)
		}
	}
	return nil
}
