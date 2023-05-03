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
	"sort"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func NewListContextsCmd(_ *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list configured contexts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			contextNames, err := getContextNames()
			if err != nil {
				return err
			}
			current, err := GetCurrentContext()
			if err != nil {
				return err
			}
			for _, name := range contextNames {
				if name == current {
					fmt.Printf("* %s\n", name)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}

			return nil
		},
	}

	return cmd
}

// getContextNames prints the list of context names, with the current context
// marked with a star (*)
func getContextNames() ([]string, error) {
	contextMap, err := getContextMap()
	if err != nil {
		return nil, err
	}
	var contextList []string
	for contextName := range contextMap {
		contextList = append(contextList, contextName)
	}
	sort.Strings(contextList)
	return contextList, nil
}
