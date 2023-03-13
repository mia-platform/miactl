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

package clioptions

import (
	"github.com/spf13/cobra"
)

type CLIOptions struct {
	CfgFile         string
	Verbose         bool
	APIKey          string
	APICookie       string
	APIToken        string
	SkipCertificate bool
	CACert          string
	Context         string
	ProjectID       string
	CompanyID       string
	APIBaseURL      string
}

func NewCLIOptions() *CLIOptions {
	return &CLIOptions{}
}

func (f *CLIOptions) AddRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "whether to output details in verbose mode")
}

func (f *CLIOptions) AddConnectionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.APIKey, "apiKey", "", "API Key")
	cmd.PersistentFlags().StringVar(&f.APICookie, "apiCookie", "", "api cookie sid")
	cmd.PersistentFlags().StringVar(&f.APIToken, "apiToken", "", "api access token")
	cmd.PersistentFlags().StringVar(&f.Context, "context", "", "The name of the context to use")
	cmd.PersistentFlags().BoolVar(&f.SkipCertificate, "insecure", false, "whether to not check server certificate")
}

func (f *CLIOptions) AddContextFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
	cmd.Flags().StringVar(&f.APIBaseURL, "endpoint", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	cmd.Flags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
	cmd.Flags().StringVar(
		&f.CACert,
		"ca-cert",
		"",
		"file path to a CA certificate, which can be employed to verify server certificate",
	)
}
