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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mia-platform/miactl/internal/cliconfig"
	"github.com/mia-platform/miactl/internal/client"
	"github.com/spf13/pflag"
)

type CLIOptions struct {
	MiactlConfig string
	Verbose      bool

	Endpoint string
	Insecure bool
	CAFile   string

	Context   string
	Auth      string
	ProjectID string
	CompanyID string

	Revision   string
	DeployType string
	NoSemVer   bool

	BasicClientID     string
	BasicClientSecret string
	JWTJsonPath       string

	ServiceAccountRole string
	OutputPath         string
}

// NewCLIOptions return a new CLIOptions instance
func NewCLIOptions() *CLIOptions {
	return &CLIOptions{}
}

func (o *CLIOptions) AddGlobalFlags(flags *pflag.FlagSet) {
	locator := cliconfig.NewConfigPathLocator()
	configFilePathDescription := fmt.Sprintf("path to the config file default to %s", locator.DefaultConfigPath())
	flags.StringVarP(&o.MiactlConfig, "config", "c", "", configFilePathDescription)
	flags.BoolVar(&o.Verbose, "verbose", false, "increase the verbosity of the cli output")
}

func (o *CLIOptions) AddConnectionFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.CAFile,
		"certificate-authority",
		o.CAFile,
		"path to a cert file for the certificate authority for the selected endpoint",
	)
	flags.StringVar(&o.Endpoint, "endpoint", "", "the address and port of the Mia-Platform Console server")
	flags.BoolVar(&o.Insecure,
		"insecure-skip-tls-verify",
		false,
		"if true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
}

func (o *CLIOptions) AddContextFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Context, "context", "", "the name of the miactl context to use")
	o.AddAuthFlags(flags)
}

func (o *CLIOptions) AddAuthFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Auth, "auth-name", "", "the name of the miactl auth to use")
}

func (o *CLIOptions) AddProjectFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.ProjectID, "project-id", "", "the ID of the project")
}

func (o *CLIOptions) AddCompanyFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.CompanyID, "company-id", "", "the ID of the company")
}

func (o *CLIOptions) AddDeployFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Revision, "revision", "HEAD", "revision of the commit to deploy")
	flags.StringVar(&o.DeployType, "deploy-type", "smart_deploy", "deploy type")
	flags.BoolVar(&o.NoSemVer, "no-semver", false, "force the deploy wihout semver")
}

func (o *CLIOptions) AddContextAuthFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.BasicClientID, "client-id", "", "the client ID of the service account")
	flags.StringVar(&o.BasicClientSecret, "client-secret", "", "the client secret of the service account")
	flags.StringVar(&o.JWTJsonPath, "jwt-json", "", "path of the json containing the json config of a jwt service account")
}

func (o *CLIOptions) AddServiceAccountFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.ServiceAccountRole, "role", "r", "", "the company role of the service account")
}

func (o *CLIOptions) AddJWTServiceAccountFlags(flags *pflag.FlagSet) {
	o.AddServiceAccountFlags(flags)
	flags.StringVarP(&o.OutputPath, "output", "o", "", "write the service account to a file")
}

func (o *CLIOptions) ToRESTConfig() (*client.Config, error) {
	locator := cliconfig.NewConfigPathLocator()
	locator.ExplicitPath = o.MiactlConfig

	config, err := locator.ReadConfig()
	if err != nil {
		return nil, err
	}

	overrides := new(cliconfig.ConfigOverrides)
	overrides.Endpoint = o.Endpoint
	overrides.CompanyID = o.CompanyID
	overrides.ProjectID = o.ProjectID
	overrides.Context = o.Context
	overrides.AuthName = o.Auth
	overrides.CertificateAuthority = o.CAFile
	overrides.InsecureSkipTLSVerify = o.Insecure

	clientConfig, err := cliconfig.NewConfigReader(config, overrides).ClientConfig(locator)
	if err != nil {
		return nil, err
	}
	clientConfig.UserAgent = defaultUserAgent()
	return clientConfig, nil
}

func defaultUserAgent() string {
	osCommand := os.Args[0]
	command := "unknown"
	if len(osCommand) > 0 {
		command = filepath.Base(osCommand)
	}

	os := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("%s (%s/%s)", command, os, arch)
}
