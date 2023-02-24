package clioptions

import "github.com/spf13/cobra"

type ConnectionOptions struct {
	APIKey                string
	APICookie             string
	APIToken              string
	SkipCertificate       bool
	AdditionalCertificate string
	Context               string
}

func (f *ConnectionOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.APIKey, "apiKey", "", "API Key")
	cmd.PersistentFlags().StringVar(&f.APICookie, "apiCookie", "", "api cookie sid")
	cmd.PersistentFlags().StringVar(&f.APIToken, "apiToken", "", "api access token")
	cmd.PersistentFlags().StringVar(&f.Context, "context", "", "The name of the context to use")
	cmd.PersistentFlags().BoolVar(&f.SkipCertificate, "insecure", false, "whether to not check server certificate")
	cmd.PersistentFlags().StringVar(
		&f.AdditionalCertificate,
		"ca-cert",
		"",
		"file path to additional CA certificate, which can be employed to verify server certificate",
	)
}

func NewConnectionOptions() *ConnectionOptions {
	return &ConnectionOptions{}
}
