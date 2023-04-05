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
	CfgFile             string
	Verbose             bool
	SkipCertificate     bool
	CACert              string
	Context             string
	ProjectID           string
	CompanyID           string
	Endpoint            string
	Revision            string
	DeployType          string
	ForceDeployNoSemVer bool
}

func NewCLIOptions() *CLIOptions {
	return &CLIOptions{}
}

func (f *CLIOptions) AddGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.CfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
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
	cmd.PersistentFlags().BoolVar(&f.SkipCertificate, "insecure", false, "whether to not check server certificate")
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
	cmd.PersistentFlags().StringVar(&f.Revision, "revision", "master", "Revision of the commit to deploy")
	cmd.PersistentFlags().StringVar(&f.DeployType, "deploy-type", "smart_deploy", "Deploy type")
	cmd.PersistentFlags().BoolVar(&f.ForceDeployNoSemVer, "forcedeploynosemver", false, "Force the deploy wihout semver")
}
