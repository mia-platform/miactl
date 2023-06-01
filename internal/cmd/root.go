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
	"github.com/mia-platform/miactl/internal/cmd/deploy"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "miactl",
		Short: "Mia-Platform Console CLI",
		Long: `miactl is a CLI for interacting with Mia-Platform Console

	Find more information at: https://docs.mia-platform.eu/docs/cli/miactl/overview`,
	}

	// initialize clioptions and setup during initialization
	options := clioptions.NewCLIOptions()

	// add cmd flags
	options.AddGlobalFlags(rootCmd.PersistentFlags())

	// add sub commands
	rootCmd.AddCommand(
		deploy.NewDeployCmd(options),
		CompanyCmd(options),
		ContextCmd(options),
		ProjectCmd(options),
		ServiceAccountCmd(options),
	)

	return rootCmd
}
