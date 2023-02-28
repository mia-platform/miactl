package clioptions

import "github.com/spf13/cobra"

type RootOptions struct {
	CfgFile string
	Verbose bool
}

func (f *RootOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "whether to output details in verbose mode")
}

func NewRootOptions() *RootOptions {
	return &RootOptions{}
}
