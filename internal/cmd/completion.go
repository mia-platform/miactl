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
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
func newCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	longUsage := `To load completion run

	For bash:

	miactl completion bash >/etc/bash_completion.d/miactl

	To configure your bash shell to load completions for each session add to your bashrc, run

	echo 'source <(miactl completion bash)' >>~/.bashrc

	For fish:

	miactl completion fish >~/.config/fish/completions/miactl.fish

	For zsh:

	To generate the completion script, run miactl completion zsh
	the generated completion script should be put somewhere in your $fpath named _miactl

	---

	After reloading your shell, miactl autocompletion should be working.
	`

	validArgs := []string{"bash", "fish", "zsh"}

	var completionCmd = &cobra.Command{
		Use:       "completion",
		Short:     "Generates bash completion scripts",
		Long:      longUsage,
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]

			switch shell {
			case "bash":
				return rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			}

			return nil
		},
	}

	return completionCmd
}
