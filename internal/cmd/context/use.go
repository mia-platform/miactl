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

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewUseContextCmd(opts *clioptions.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use CONTEXT [flags]",
		Short: "update available contexts for miactl",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := contextLookUp(args[0]); err != nil {
				return fmt.Errorf("error looking up the context in the config file: %w", err)
			}
			viper.Set("current-context", args[0])
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("error updating the configuration: %w", err)
			}
			fmt.Println("OK")
			return nil
		},
	}

	return cmd
}

func contextLookUp(contextName string) error {
	if viper.Get("contexts") == nil {
		return fmt.Errorf("no context specified in config file")
	}
	contextMap := viper.Get("contexts").(map[string]interface{})
	if contextMap[contextName] == nil {
		return fmt.Errorf("context %s does not exist", contextName)
	}
	return nil
}
