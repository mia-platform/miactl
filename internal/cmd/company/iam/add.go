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

package iam

import (
	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/company/iam/group"
	"github.com/mia-platform/miactl/internal/cmd/company/iam/serviceaccount"
	"github.com/mia-platform/miactl/internal/cmd/company/iam/user"
)

func AddCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new IAM entity to a company",
		Long: `A Company can have associated different entities for managing the roles, this command will create or
add them to the company selected via the flag or context`,
	}

	cmd.AddCommand(
		serviceaccount.CreateServiceAccountCmd(options),
		user.AddCmd(options),
		group.AddCmd(options),
		group.AddMemberCmd(options),
	)

	return cmd
}
