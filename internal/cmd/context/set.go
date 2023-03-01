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

func NewSetContextCmd(rootOptions *clioptions.RootOptions) *cobra.Command {
	contextOptions := clioptions.NewContextOptions(rootOptions)
	cmd := &cobra.Command{
		Use:   "set CONTEXT [flags]",
		Short: "update available contexts for miactl",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updatedContexts := updateContextMap(contextOptions, args[0])
			viper.Set("contexts", updatedContexts)
			if err := viper.WriteConfig(); err != nil {
				fmt.Println("error saving the configuration")
				return err
			}
			fmt.Println("OK")
			return nil
		},
	}
	contextOptions.AddContextFlags(cmd)
	return cmd
}

func updateContextMap(opts *clioptions.ContextOptions, contextName string) map[string]interface{} {
	contextMap := make(map[string]interface{})
	if viper.Get("contexts") != nil {
		contextMap = viper.Get("contexts").(map[string]interface{})
	}
	if contextMap[contextName] == nil {
		newContext := map[string]string{"apibaseurl": opts.APIBaseURL, "projectid": opts.ProjectID, "companyid": opts.CompanyID}
		contextMap[contextName] = newContext
	} else {
		oldContext := contextMap[contextName].(map[string]interface{})
		if opts.APIBaseURL != "https://console.cloud.mia-platform.eu" {
			oldContext["apibaseurl"] = opts.APIBaseURL
		}
		if opts.ProjectID != "" {
			oldContext["projectid"] = opts.ProjectID
		}
		if opts.CompanyID != "" {
			oldContext["companyid"] = opts.CompanyID
		}
	}
	return contextMap
}
