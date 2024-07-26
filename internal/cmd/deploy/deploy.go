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

package deploy

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

func NewDeployCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy ENVIRONMENT",
		Short: "Deploy the target environment.",
		Long: `Deprecation Warning: This command is deprecated. Use 'deploy trigger' instead.

Trigger the deploy of the target environment in the selected project.

The deploy will be performed by the pipeline setup in project, the command will then keep
listening on updates of the status for keep the user informed on the updates. The command
will exit with error if the pipeline will not end with a success.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Deprecation Warning: The 'deploy' command is deprecated. Use 'deploy trigger' instead.\n")

			environmentName := args[0]
			return runDeployTrigger(cmd.Context(), environmentName, options)
		},
	}

	deployTriggerOptions(cmd, options)

	cmd.AddCommand(
		triggerCmd(options),
		addCmd(options),
	)

	return cmd
}

func addCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "add",
	}

	cmd.AddCommand(newStatusAddCmd(options))

	return cmd
}
