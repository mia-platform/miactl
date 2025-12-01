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

package company

import (
	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/company/iam"
)

func IAMCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage Mia-Platform IAM for a company",
		Long: `A Company Owner user can manage the access to the company directly to a user,
via a group or through service accounts.`,
	}

	// add cmd flags
	flags := cmd.PersistentFlags()
	o.AddConnectionFlags(flags)
	o.AddContextFlags(flags)
	o.AddCompanyFlags(flags)

	cmd.AddCommand(
		iam.ListCmd(o),
		iam.AddCmd(o),
		iam.EditCmd(o),
		iam.RemoveCmd(o),
	)

	return cmd
}
