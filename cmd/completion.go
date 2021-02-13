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
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {

			shell := args[0]

			switch shell {
			case "bash":
				rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "fish":
				rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "zsh":
				rootCmd.GenZshCompletion(cmd.OutOrStdout())
			}
		},
	}

	return completionCmd
}
