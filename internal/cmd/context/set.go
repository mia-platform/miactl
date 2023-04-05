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

func NewSetContextCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set CONTEXT [flags]",
		Short: "update available contexts for miactl",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updatedContexts := updateContextMap(options, args[0])
			viper.Set("contexts", updatedContexts)
			if err := viper.WriteConfig(); err != nil {
				fmt.Println("error saving the configuration")
				return err
			}
			fmt.Println("OK")
			return nil
		},
	}
	options.AddConnectionFlags(cmd)
	options.AddCompanyFlags(cmd)
	options.AddProjectFlags(cmd)
	return cmd
}

func updateContextMap(opts *clioptions.CLIOptions, contextName string) map[string]interface{} {
	contextMap := make(map[string]interface{})
	if viper.Get("contexts") != nil {
		contextMap = viper.Get("contexts").(map[string]interface{})
	}
	newContext := make(map[string]interface{})
	newContext["endpoint"] = opts.Endpoint
	if opts.ProjectID != "" {
		newContext["projectid"] = opts.ProjectID
	}
	if opts.CompanyID != "" {
		newContext["companyid"] = opts.CompanyID
	}
	if opts.CACert != "" {
		newContext["ca-cert"] = opts.CACert
	}
	if opts.Insecure {
		newContext["insecure"] = opts.Insecure
	}

	contextMap[contextName] = newContext

	return contextMap
}
