package clioptions

import "github.com/spf13/cobra"

type RootOptions struct {
	ProjectID  string
	CompanyID  string
	CfgFile    string
	Verbose    bool
	APIBaseURL string
}

func (f *RootOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
	cmd.PersistentFlags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
	cmd.PersistentFlags().StringVar(&f.APIBaseURL, "endpoint", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	cmd.PersistentFlags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "whether to output details in verbose mode")
}

func NewRootOptions() *RootOptions {
	return &RootOptions{}
}
