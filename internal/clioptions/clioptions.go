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
	"github.com/mia-platform/miactl/internal/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CLIOptions struct {
	MiactlConfig string

	Endpoint string
	Insecure bool
	CAFile   string

	Context     string
	Auth        string
	ProjectID   string
	CompanyID   string
	Environment string

	Revision   string
	DeployType string
	NoSemVer   bool

	IAMRole            string
	ProjectIAMRole     string
	EnvironmentIAMRole string
	EntityID           string

	UserEmail                 string
	KeepUserGroupMemeberships bool

	UserEmails []string
	UserIDs    []string

	ServiceAccountID string

	BasicClientID     string
	BasicClientSecret string
	JWTJsonPath       string
	OutputPath        string

	InputFilePath string

	MarketplaceResourcePaths []string
	// MarketplaceItemID is the itemId field of a Marketplace item
	MarketplaceItemID string
	// MarketplaceItemVersion is the version field of a Marketplace item
	MarketplaceItemVersion string
	// MarketplaceItemObjectID is the _id of a Marketplace item
	MarketplaceItemObjectID     string
	MarketplaceFetchPublicItems bool

	FromCronJob string

	FollowLogs bool

	// OutputFormat describes the output format of some commands. Can be json or yaml.
	OutputFormat string

	ShowUsers           bool
	ShowGroups          bool
	ShowServiceAccounts bool
}

// NewCLIOptions return a new CLIOptions instance
func NewCLIOptions() *CLIOptions {
	return &CLIOptions{}
}

func (o *CLIOptions) AddGlobalFlags(flags *pflag.FlagSet) {
	locator := cliconfig.NewConfigPathLocator()
	configFilePathDescription := fmt.Sprintf("path to the config file default to %s", locator.DefaultConfigPath())
	flags.StringVarP(&o.MiactlConfig, "config", "c", "", configFilePathDescription)
	flags.IntVarP(&logger.LogLevel, "verbose", "v", 0, "increase the verbosity of the cli output")
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

func (o *CLIOptions) AddEnvironmentFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Environment, "environment", "", "the environment scope for the command")
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
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the company role of the service account")
}

func (o *CLIOptions) AddJWTServiceAccountFlags(flags *pflag.FlagSet) {
	o.AddServiceAccountFlags(flags)
	flags.StringVarP(&o.OutputPath, "output", "o", "", "write the service account configuration as json to a file")
}

func (o *CLIOptions) AddEditServiceAccountFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the new company role for the service account")
	flags.StringVar(&o.ServiceAccountID, "service-account-id", "", "the service account id to edit")
}

func (o *CLIOptions) AddRemoveServiceAccountFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.ServiceAccountID, "service-account-id", "", "the service account id to remove")
}

func (o *CLIOptions) AddNewUserFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the company role of the user")
	flags.StringVar(&o.UserEmail, "email", "", "the email of the user to add")
}

func (o *CLIOptions) AddEditUserFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the new company role for the user")
	flags.StringVar(&o.EntityID, "user-id", "", "the user id to edit")
}

func (o *CLIOptions) AddRemoveUserFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.EntityID, "user-id", "", "the user id to remove")
	flags.BoolVar(&o.KeepUserGroupMemeberships, "no-include-groups", false, "keep the user membership in the company groups")
}

func (o *CLIOptions) CreateNewGroupFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the company role of the group")
}

func (o *CLIOptions) AddNewMembersToGroupFlags(flags *pflag.FlagSet) {
	flags.StringSliceVar(&o.UserEmails, "user-email", []string{}, "the list of user email to add to the group")
	flags.StringVar(&o.EntityID, "group-id", "", "the group id where to add the users")
}

func (o *CLIOptions) AddEditGroupFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.IAMRole, "role", "r", "", "the new company role for the group")
	flags.StringVar(&o.EntityID, "group-id", "", "the group id to edit")
}

func (o *CLIOptions) AddRemoveGroupFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.EntityID, "group-id", "", "the group id to remove")
}

func (o *CLIOptions) AddRemoveMembersFromGroupFlags(flags *pflag.FlagSet) {
	flags.StringSliceVar(&o.UserIDs, "user-id", []string{}, "the list of user id to remove to the group")
	flags.StringVar(&o.EntityID, "group-id", "", "the group id where to remove the users")
}

func (o *CLIOptions) AddMarketplaceApplyFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&o.MarketplaceResourcePaths, "file-path", "f", []string{}, "paths to JSON/YAML files or folder of files containing a Marketplace item definition")
	err := cmd.MarkFlagRequired("file-path")
	if err != nil {
		// the error is only due to a programming error (missing command), hence panic
		panic(err)
	}
}

func (o *CLIOptions) AddMarketplaceItemIDFlag(flags *pflag.FlagSet) (flagName string) {
	flagName = "item-id"
	flags.StringVarP(&o.MarketplaceItemID, flagName, "i", "", "The itemId of the Marketplace item")
	return
}

func (o *CLIOptions) AddPublicFlag(flags *pflag.FlagSet) (flagName string) {
	flagName = "public"
	flags.BoolVarP(&o.MarketplaceFetchPublicItems, flagName, "p", false, "specify to fetch also public items")
	return
}

func (o *CLIOptions) AddMarketplaceItemObjectIDFlag(flags *pflag.FlagSet) (flagName string) {
	flagName = "object-id"
	flags.StringVar(&o.MarketplaceItemObjectID, flagName, "", "The _id of the Marketplace item")
	return
}

func (o *CLIOptions) AddMarketplaceVersionFlag(flags *pflag.FlagSet) (flagName string) {
	flagName = "version"
	flags.StringVar(&o.MarketplaceItemVersion, flagName, "", "The version of the Marketplace item")
	return
}

func (o *CLIOptions) AddCreateJobFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.FromCronJob, "from", "", "name of the cronjob to create a Job from")
}

func (o *CLIOptions) AddLogsFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.FollowLogs, "follow", "f", false, "specify if the logs should be streamed")
}

func (o *CLIOptions) AddOutputFormatFlag(flags *pflag.FlagSet, defaultVal string) {
	flags.StringVarP(&o.OutputFormat, "output", "o", defaultVal, "Output format. Allowed values: json, yaml")
}

func (o *CLIOptions) AddIAMListFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&o.ShowUsers, "users", false, "Filter IAM entities to show only users. Mutally exclusive with groups and serviceAccounts")
	flags.BoolVar(&o.ShowGroups, "groups", false, "Filter IAM entities to show only groups. Mutally exclusive with users and serviceAccounts")
	flags.BoolVar(&o.ShowServiceAccounts, "serviceAccounts", false, "Filter IAM entities to show only service accounts. Mutally exclusive with users and groups")
}

func (o *CLIOptions) AddEditCompanyIAMFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.EntityID, "entity-id", "", "the entity id to change")
	flags.StringVar(&o.ProjectIAMRole, "project-role", "", "the new role for the current project")
	flags.StringVar(&o.EnvironmentIAMRole, "environment-role", "", "the new role for the selected environment")
	flags.StringVar(&o.Environment, "environment", "", "the environment where to change the role")
}

func (o *CLIOptions) AddRemoveProjectIAMRoleFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.EntityID, "entity-id", "", "the entity id to change")
	flags.StringVar(&o.Environment, "environment", "", "set the flag to the environment name for deleting the role for that environment")
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
	overrides.Environment = o.Environment
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
