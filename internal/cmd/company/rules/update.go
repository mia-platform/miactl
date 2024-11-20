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

package rules

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/files"
	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"
	"github.com/spf13/cobra"
)

func UpdateRules(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update rules",
		Long:  "Update company or project rules from file  (this command is related to a closed preview feature, it may be subject to breaking changes!)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" && restConfig.ProjectID == "" {
				return ErrRequiredCompanyIDOrProjectID
			}

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			rules, err := readFile(o.InputFilePath)
			cobra.CheckErr(err)

			if restConfig.ProjectID != "" {
				err = New(client).UpdateProjectRules(cmd.Context(), restConfig.ProjectID, rules)
				cobra.CheckErr(err)
			} else {
				err = New(client).UpdateTenantRules(cmd.Context(), restConfig.CompanyID, rules)
				cobra.CheckErr(err)
			}

			fmt.Printf("Rules updated successfully")
			return nil
		},
	}

	requireFilePathFlag(o, cmd)

	return cmd
}

func requireFilePathFlag(o *clioptions.CLIOptions, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.InputFilePath, "file-path", "f", "", "paths to JSON file containing the ruleset definition")
	err := cmd.MarkFlagRequired("file-path")
	if err != nil {
		panic(err)
	}
}

func readFile(path string) ([]*rulesentities.SaveChangesRules, error) {
	data := []*rulesentities.SaveChangesRules{}
	if err := files.ReadFile(path, &data); err != nil {
		return data, err
	}

	return data, nil
}
