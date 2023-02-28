package clioptions

import "github.com/spf13/cobra"

type ContextOptions struct {
	RootOptions
	ProjectID  string
	CompanyID  string
	APIBaseURL string
}

func (f *ContextOptions) AddContextFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
	cmd.Flags().StringVar(&f.APIBaseURL, "endpoint", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	cmd.Flags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
}

func NewContextOptions(rootOptions *RootOptions) *ContextOptions {
	return &ContextOptions{
		RootOptions: *rootOptions,
	}
}
