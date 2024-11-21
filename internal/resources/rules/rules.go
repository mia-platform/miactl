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

type SaveChangesRules struct {
	RoleIDs           []string  `yaml:"roleIds,omitempty" json:"roleIds,omitempty"` //nolint:tagliatelle
	DisallowedRuleSet []RuleSet `yaml:"disallowedRuleSet,omitempty" json:"disallowedRuleSet,omitempty"`
}

type ProjectSaveChangesRules struct {
	RoleIDs               []string  `yaml:"roleIds,omitempty" json:"roleIds,omitempty"` //nolint:tagliatelle
	DisallowedRuleSet     []RuleSet `yaml:"disallowedRuleSet,omitempty" json:"disallowedRuleSet,omitempty"`
	IsInheritedFromTenant bool      `yaml:"isInheritedFromTenant,omitempty" json:"isInheritedFromTenant,omitempty"`
}

type RuleSet struct {
	JSONPath string       `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Options  *RuleOptions `yaml:"processingOptions,omitempty" json:"processingOptions,omitempty"` //nolint:tagliatelle
	RuleID   string       `yaml:"ruleId,omitempty" json:"ruleId,omitempty"`                       //nolint:tagliatelle
}

type RuleOptions struct {
	Action     string `yaml:"action,omitempty" json:"action,omitempty"`
	PrimaryKey string `yaml:"primaryKey,omitempty" json:"primaryKey,omitempty"`
}
