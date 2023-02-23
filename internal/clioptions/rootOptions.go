package clioptions

import "github.com/spf13/cobra"

type RootOptions struct {
	APIKey                string
	APICookie             string
	APIBaseURL            string
	APIToken              string
	SkipCertificate       bool
	AdditionalCertificate string
	ProjectID             string
	CompanyID             string
	CfgFile               string
	Context               string
	Verbose               bool
}

func (f *RootOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
	cmd.PersistentFlags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
	cmd.PersistentFlags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
	cmd.PersistentFlags().StringVar(&f.APIKey, "apiKey", "", "API Key")
	cmd.PersistentFlags().StringVar(&f.APICookie, "apiCookie", "", "api cookie sid")
	cmd.PersistentFlags().StringVar(&f.APIBaseURL, "endpoint", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	cmd.PersistentFlags().StringVar(&f.APIToken, "apiToken", "", "api access token")
	cmd.PersistentFlags().StringVar(&f.Context, "context", "", "The name of the context to use")
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "whether to output details in verbose mode")
	cmd.PersistentFlags().BoolVar(&f.SkipCertificate, "insecure", false, "whether to not check server certificate")
	cmd.PersistentFlags().StringVar(
		&f.AdditionalCertificate,
		"ca-cert",
		"",
		"file path to additional CA certificate, which can be employed to verify server certificate",
	)
}

func NewRootOptions() *RootOptions {
	return &RootOptions{}
}
