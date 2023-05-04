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
	CfgFile           string
	Verbose           bool
	Insecure          bool
	CACert            string
	Context           string
	ProjectID         string
	CompanyID         string
	Endpoint          string
	Revision          string
	DeployType        string
	NoSemVer          bool
	BasicClientID     string
	BasicClientSecret string
	AuthMethod        string
	CompanyRole       string
}

func NewCLIOptions() *CLIOptions {
	return &CLIOptions{}
}

func (f *CLIOptions) AddGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config)")
	cmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "whether to output details in verbose mode")
}

func (f *CLIOptions) AddConnectionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&f.CACert,
		"ca-cert",
		"",
		"file path to a CA certificate, which can be employed to verify server certificate",
	)
	cmd.PersistentFlags().StringVar(&f.Endpoint, "endpoint", "https://console.cloud.mia-platform.eu", "The URL of the console endpoint")
	cmd.PersistentFlags().BoolVar(&f.Insecure, "insecure", false, "whether to not check server certificate")
}

func (f *CLIOptions) AddContextFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.Context, "context", "", "The name of the context to use")
}

func (f *CLIOptions) AddProjectFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.ProjectID, "project-id", "", "The ID of the project")
}

func (f *CLIOptions) AddCompanyFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CompanyID, "company-id", "", "The ID of the company")
}

func (f *CLIOptions) AddDeployFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.Revision, "revision", "", "Revision of the commit to deploy")
	if err := cmd.MarkPersistentFlagRequired("revision"); err != nil {
		// if there is an error something very wrong is happening, panic
		panic(err)
	}
	cmd.PersistentFlags().StringVar(&f.DeployType, "deploy-type", "smart_deploy", "Deploy type")
	cmd.PersistentFlags().BoolVar(&f.NoSemVer, "no-semver", false, "Force the deploy wihout semver")
}

func (f *CLIOptions) AddBasicAuthFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.BasicClientID, "client-id", "", "The client ID of the service account")
	cmd.PersistentFlags().StringVar(&f.BasicClientSecret, "client-secret", "", "The client secret of the service account")
}

func (f *CLIOptions) AddServiceAccountFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.AuthMethod, "method", "", "The authentication method of the service account")
	cmd.PersistentFlags().StringVar(&f.CompanyRole, "role", "", "The company role of the service account")
}
