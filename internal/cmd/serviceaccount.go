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

package cmd

import (
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/serviceaccount"
	"github.com/spf13/cobra"
)

func ServiceAccountCmd(o *clioptions.CLIOptions) *cobra.Command {
	serviceAccountCmd := &cobra.Command{
		Use:   "serviceaccount",
		Short: "Manage Mia-Platform Console service accounts",
		Long: `Manage service accounts for connecting to Mia-Platform Console.

Service Account credentials can be useful for setting a login with different
permissions or enable authentication in scripts, automations and pipelines.
Mia-Platform Console support two kinds of authentication: basic or jwt.`,
	}

	// add cmd flags
	o.AddConnectionFlags(serviceAccountCmd.PersistentFlags())
	o.AddContextFlags(serviceAccountCmd.PersistentFlags())
	o.AddCompanyFlags(serviceAccountCmd.PersistentFlags())

	// add sub commands
	serviceAccountCmd.AddCommand(serviceaccount.CreateServiceAccountCmd(o))

	return serviceAccountCmd
}
