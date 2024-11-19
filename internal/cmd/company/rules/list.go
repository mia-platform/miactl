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
	"strconv"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"
	"github.com/spf13/cobra"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured rules",
		Long:  "List all the rules configured for the specified company",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" {
				return ErrRequiredCompanyID
			}

			rules, err := New(client).ListTenantRules(cmd.Context(), restConfig.CompanyID)
			cobra.CheckErr(err)

			printList(rules, options.Printer(clioptions.DisableWrapLines(true)))
			return nil
		},
	}

	return cmd
}

func printList(rules []*rulesentities.SaveChangesRules, p printer.IPrinter) {
	tableColumnLabel := []string{"#", "Roles", "Ruleset"}

	p.Keys(tableColumnLabel...)
	for i, rule := range rules {

		ruleInfo := []string{}
		for _, ruleset := range rule.DisallowedRuleSet {
			if ruleset.RuleID != "" {
				ruleInfo = append(ruleInfo, fmt.Sprintf("Rule ID: '%s'", ruleset.RuleID))
				continue
			}
			if ruleset.JSONPath != "" {
				ruleInfo = append(ruleInfo, fmt.Sprintf("JSON Path: '%s'", ruleset.JSONPath))
				continue
			}
		}

		p.Record(
			strconv.Itoa(i),
			strings.Join(rule.RoleIDs, ", "),
			strings.Join(ruleInfo, ", "),
		)
	}
	p.Print()
}
