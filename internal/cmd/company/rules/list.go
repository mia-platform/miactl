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

var (
	ErrRequiredCompanyIDOrProjectID = fmt.Errorf("at least one of company id or project id is required, please set it via flag or context")
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured rules",
		Long:  "List all the rules configured for the specified company or project (this command is related to a closed preview feature, it may be subject to breaking changes!)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			if restConfig.CompanyID == "" && restConfig.ProjectID == "" {
				return ErrRequiredCompanyIDOrProjectID
			}

			if restConfig.ProjectID != "" {
				rules, err := New(client).ListProjectRules(cmd.Context(), restConfig.ProjectID)
				cobra.CheckErr(err)
				printProjectList(rules, options.Printer(clioptions.DisableWrapLines(true)))
				return nil
			}

			rules, err := New(client).ListTenantRules(cmd.Context(), restConfig.CompanyID)
			cobra.CheckErr(err)
			printTenantList(rules, options.Printer(clioptions.DisableWrapLines(true)))
			return nil
		},
	}

	return cmd
}

func createRecord(disallowedRules []rulesentities.RuleSet, allowedRules []rulesentities.RuleSet) []string {
	ruleInfo := []string{}
	for _, dRuleset := range disallowedRules {
		if dRuleset.RuleID != "" {
			ruleInfo = append(ruleInfo, fmt.Sprintf("Disallowed Rule ID: '%s'", dRuleset.RuleID))
			continue
		}
		if dRuleset.JSONPath != "" {
			ruleInfo = append(ruleInfo, fmt.Sprintf("Disallowed JSON Path: '%s'", dRuleset.JSONPath))
			continue
		}
	}
	for _, aRuleset := range allowedRules {
		if aRuleset.RuleID != "" {
			ruleInfo = append(ruleInfo, fmt.Sprintf("Allowed Rule ID: '%s'", aRuleset.RuleID))
			continue
		}
		if aRuleset.JSONPath != "" {
			ruleInfo = append(ruleInfo, fmt.Sprintf("Allowed JSON Path: '%s'", aRuleset.JSONPath))
			continue
		}
	}
	return ruleInfo
}

func printTenantList(rules []*rulesentities.SaveChangesRules, p printer.IPrinter) {
	tableColumnLabel := []string{"#", "Roles", "Ruleset"}

	p.Keys(tableColumnLabel...)
	for i, rule := range rules {
		ruleInfo := createRecord(rule.DisallowedRuleSet, rule.AllowedRuleSet)
		p.Record(
			strconv.Itoa(i),
			strings.Join(rule.RoleIDs, ", "),
			strings.Join(ruleInfo, ", "),
		)
	}
	p.Print()
}

func printProjectList(rules []*rulesentities.ProjectSaveChangesRules, p printer.IPrinter) {
	tableColumnLabel := []string{"#", "Roles", "Ruleset", "Inherited"}

	p.Keys(tableColumnLabel...)
	for i, rule := range rules {
		ruleInfo := createRecord(rule.DisallowedRuleSet, rule.AllowedRuleSet)
		p.Record(
			strconv.Itoa(i),
			strings.Join(rule.RoleIDs, ", "),
			strings.Join(ruleInfo, ", "),
			strconv.FormatBool(rule.IsInheritedFromTenant),
		)
	}
	p.Print()
}
