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

package project

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	importConfigCmdUsage = "import-config"
	importConfigCmdShort = "[beta] - Import configuration from JSON files"
	importConfigCmdLong  = `[beta] - Import configuration from JSON files in a Mia-Platform Console project.
This command allows you to import configurations from specific JSON files:
- flowManagerConfig.json
- rbacManagerConfig.json
- backofficeConfigurations.json
- fast-data-config.json
- api-console-config.json
`
)

const (
	configRevisionEndpointTemplate    = "/api/backend/projects/%s/revisions/%s/configuration"
	configEnvironmentEndpointTemplate = "/api/projects/%s/environments/%s/configuration"
)

// ConfigImportOptions contains the options for configuration import
type ConfigImportOptions struct {
	FlowManagerConfigPath string
	RbacManagerConfigPath string
	BackofficeConfigPath  string
	FastDataConfigPath    string
	APIConsoleConfigPath  string
	Environment           string // Environment ID for saving to environment configuration
	SkipConfirmation      bool   // If true, skip interactive confirmation
}

// ImportConfigCmd returns the command for importing configurations
func ImportConfigCmd(o *clioptions.CLIOptions) *cobra.Command {
	importOptions := &ConfigImportOptions{}

	cmd := &cobra.Command{
		Use:   importConfigCmdUsage,
		Short: importConfigCmdShort,
		Long:  importConfigCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return importConfigFromFiles(cmd.Context(), client, restConfig.ProjectID, o.Revision, importOptions, cmd.ErrOrStderr())
		},
	}

	flags := cmd.Flags()
	o.AddProjectFlags(flags)
	flags.StringVar(&o.Revision, "revision", "", "revision of the commit where to import the configuration (mutually exclusive with --environment)")
	flags.StringVar(&importOptions.Environment, "environment", "", "environment where to import the configuration (mutually exclusive with --revision)")

	flags.StringVar(&importOptions.FlowManagerConfigPath, "flow-manager-config", "", "path to flowManagerConfig.json file")
	flags.StringVar(&importOptions.RbacManagerConfigPath, "rbac-manager-config", "", "path to rbacManagerConfig.json file")
	flags.StringVar(&importOptions.BackofficeConfigPath, "backoffice-config", "", "path to backofficeConfigurations.json file")
	flags.StringVar(&importOptions.FastDataConfigPath, "fast-data-config", "", "path to fast-data-config.json file")
	flags.StringVar(&importOptions.APIConsoleConfigPath, "api-console-config", "", "path to api-console-config.json file")

	flags.BoolVarP(&importOptions.SkipConfirmation, "yes", "y", false, "skip interactive confirmation and proceed with import")

	return cmd
}

func validateImportOptions(projectID, revision string, opts *ConfigImportOptions) error {
	if err := validateProjectAndEnvironment(projectID, revision, opts.Environment); err != nil {
		return err
	}

	if err := validateAtLeastOneConfigFile(opts); err != nil {
		return err
	}

	return validateFilePaths(opts)
}

func validateProjectAndEnvironment(projectID, revision, environment string) error {
	if len(projectID) == 0 {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	if (len(revision) == 0 && len(environment) == 0) || (len(revision) > 0 && len(environment) > 0) {
		return fmt.Errorf("either revision or environment must be specified, but not both")
	}

	return nil
}

func validateAtLeastOneConfigFile(opts *ConfigImportOptions) error {
	if opts.FlowManagerConfigPath == "" &&
		opts.RbacManagerConfigPath == "" &&
		opts.BackofficeConfigPath == "" &&
		opts.FastDataConfigPath == "" &&
		opts.APIConsoleConfigPath == "" {
		return fmt.Errorf("missing configuration files, please specify at least one configuration file")
	}
	return nil
}

func getConfigurationEndpoint(projectID, revision string, environment string) string {
	if len(revision) > 0 {
		return fmt.Sprintf(configRevisionEndpointTemplate, projectID, revision)
	}
	return fmt.Sprintf(configEnvironmentEndpointTemplate, projectID, environment)
}

func fetchCurrentConfig(ctx context.Context, client *client.APIClient, endpoint string) (map[string]any, string, error) {
	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)

	if err != nil {
		return nil, "", err
	}
	if err := response.Error(); err != nil {
		return nil, "", err
	}

	projectConfig := make(map[string]any)
	if err := response.ParseResponse(&projectConfig); err != nil {
		return nil, "", fmt.Errorf("cannot parse project configuration: %w", err)
	}

	var previousSave string
	if changesID, ok := projectConfig["changesId"].(string); ok {
		previousSave = changesID
	}

	return projectConfig, previousSave, nil
}

func loadMicrofrontendConfig(filePath, configKey string) (map[string]any, error) {
	var configData map[string]any
	if err := readJSONFile(filePath, &configData); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", configKey, err)
	}

	if config, ok := configData["config"]; ok {
		return map[string]any{configKey: config}, nil
	}
	return map[string]any{configKey: configData}, nil
}

func updateMicrofrontendConfig(projectConfig map[string]any, opts *ConfigImportOptions) error {
	if projectConfig["microfrontendPluginsConfig"] == nil {
		projectConfig["microfrontendPluginsConfig"] = make(map[string]any)
	}

	microfrontendConfig, _ := projectConfig["microfrontendPluginsConfig"].(map[string]any)

	configsToLoad := []struct {
		path      string
		configKey string
	}{
		{opts.FlowManagerConfigPath, "flowManagerConfig"},
		{opts.RbacManagerConfigPath, "rbacManagerConfig"},
		{opts.BackofficeConfigPath, "backofficeConfigurations"},
	}

	for _, config := range configsToLoad {
		if config.path == "" {
			continue
		}

		if err := loadAndMergeMicrofrontendConfig(microfrontendConfig, config.path, config.configKey); err != nil {
			return err
		}
	}

	return nil
}

func loadAndMergeMicrofrontendConfig(microfrontendConfig map[string]any, filePath, configKey string) error {
	config, err := loadMicrofrontendConfig(filePath, configKey)
	if err != nil {
		return err
	}

	for k, v := range config {
		microfrontendConfig[k] = v
	}

	return nil
}

func updateFastDataConfig(projectConfig map[string]any, fastDataPath string) error {
	if fastDataPath != "" {
		var fastDataData map[string]any
		if err := readJSONFile(fastDataPath, &fastDataData); err != nil {
			return fmt.Errorf("error reading fast data config: %w", err)
		}

		if config, ok := fastDataData["config"]; ok {
			projectConfig["fastDataConfig"] = config
		} else {
			projectConfig["fastDataConfig"] = fastDataData
		}
	}

	return nil
}

func updateAPIConsoleConfig(projectConfig map[string]any, apiConsolePath string) error {
	if apiConsolePath != "" {
		var apiConsoleData map[string]any
		if err := readJSONFile(apiConsolePath, &apiConsoleData); err != nil {
			return fmt.Errorf("error reading API console config: %w", err)
		}

		var configToMerge map[string]any
		if config, ok := apiConsoleData["config"]; ok {
			configToMerge = config.(map[string]any)
		} else {
			configToMerge = apiConsoleData
		}

		for k, v := range configToMerge {
			projectConfig[k] = v
		}
	}

	return nil
}

func saveConfigurationToServer(ctx context.Context, client *client.APIClient, endpoint string, postConfig map[string]any, writer io.Writer) error {
	body, err := resources.EncodeResourceToJSON(postConfig)
	if err != nil {
		return fmt.Errorf("cannot encode project configuration: %w", err)
	}

	response, err := client.
		Post().
		APIPath(endpoint).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	fmt.Fprintln(writer, "Configuration imported successfully")
	return nil
}

// ImportConfigFromFiles imports configuration files into a Mia-Platform Console project.
// It fetches the current configuration from the specified project and revision or environment,
// updates it with the configurations from the provided files, and saves it back to the server
// after user confirmation (unless skip confirmation is enabled).
func ImportConfigFromFiles(ctx context.Context, client *client.APIClient, projectID, revision string, opts *ConfigImportOptions, writer io.Writer) error {
	endpoint, err := prepareImportContext(projectID, revision, opts)
	if err != nil {
		return err
	}

	projectConfig, previousSave, err := fetchCurrentConfig(ctx, client, endpoint)
	if err != nil {
		return err
	}

	if err := updateAllConfigurations(projectConfig, opts); err != nil {
		return err
	}

	if !shouldProceedWithImport(opts, revision, writer) {
		return nil
	}

	return finalizeConfigImport(ctx, client, endpoint, projectConfig, previousSave, writer)
}

func importConfigFromFiles(ctx context.Context, client *client.APIClient, projectID, revision string, opts *ConfigImportOptions, writer io.Writer) error {
	return ImportConfigFromFiles(ctx, client, projectID, revision, opts, writer)
}

func prepareImportContext(projectID, revision string, opts *ConfigImportOptions) (string, error) {
	if err := validateImportOptions(projectID, revision, opts); err != nil {
		return "", err
	}

	return getConfigurationEndpoint(projectID, revision, opts.Environment), nil
}

func shouldProceedWithImport(opts *ConfigImportOptions, revision string, writer io.Writer) bool {
	if opts.SkipConfirmation {
		return true
	}

	if !showConfirmationAndPrompt(opts, revision, writer) {
		fmt.Fprintln(writer, "Operation cancelled by user")
		return false
	}

	return true
}

func finalizeConfigImport(ctx context.Context, client *client.APIClient, endpoint string, projectConfig map[string]any, previousSave string, writer io.Writer) error {
	postConfig := PreparePostConfig(projectConfig, previousSave)
	return saveConfigurationToServer(ctx, client, endpoint, postConfig, writer)
}

func updateAllConfigurations(projectConfig map[string]any, opts *ConfigImportOptions) error {
	updateFuncs := []func() error{
		func() error {
			return updateMicrofrontendConfig(projectConfig, opts)
		},
		func() error {
			return updateFastDataConfig(projectConfig, opts.FastDataConfigPath)
		},
		func() error {
			return updateAPIConsoleConfig(projectConfig, opts.APIConsoleConfigPath)
		},
	}

	for _, updateFunc := range updateFuncs {
		if err := updateFunc(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateFilePaths checks if all specified configuration file paths exist.
// It returns an error if any of the specified files cannot be found.
func ValidateFilePaths(opts *ConfigImportOptions) error {
	filesToValidate := map[string]string{
		"flow manager config":       opts.FlowManagerConfigPath,
		"RBAC manager config":       opts.RbacManagerConfigPath,
		"backoffice configurations": opts.BackofficeConfigPath,
		"fast data config":          opts.FastDataConfigPath,
		"API console config":        opts.APIConsoleConfigPath,
	}

	for description, path := range filesToValidate {
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("%s file not found: %w", description, err)
		}
	}

	return nil
}

func validateFilePaths(opts *ConfigImportOptions) error {
	return ValidateFilePaths(opts)
}

func readJSONFile(filePath string, target interface{}) error {
	ext := filepath.Ext(filePath)
	if ext != ".json" {
		return fmt.Errorf("file must be .json: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
	}

	return nil
}

func prepareAPIConsoleConfig(config map[string]any) map[string]any {
	apiConsoleConfig := make(map[string]any)

	fieldsToExclude := map[string]bool{
		"fastDataConfig":             true,
		"microfrontendPluginsConfig": true,
		"extensionsConfig":           true,
		"committedDate":              true,
		"lastCommitAuthor":           true,
		"platformVersion":            true,
		"changesId":                  true,
	}

	for k, v := range config {
		if !fieldsToExclude[k] {
			apiConsoleConfig[k] = v
		}
	}

	return apiConsoleConfig
}

func prepareExtensionsConfig(config map[string]any) map[string]any {
	if extensions, ok := config["extensionsConfig"]; ok && extensions != nil {
		return extensions.(map[string]any)
	}
	return map[string]any{
		"files": make(map[string]any),
	}
}

func prepareMicrofrontendConfig(config map[string]any) map[string]any {
	mfeConfig := make(map[string]any)
	if mfe, ok := config["microfrontendPluginsConfig"]; ok && mfe != nil {
		mfeConfig = mfe.(map[string]any)
	}
	return mfeConfig
}

// PreparePostConfig creates the configuration structure to be sent to the server.
// It organizes the configuration into sections for API console, fast data, extensions,
// and microfrontend plugins, and includes metadata like title and previous save reference.
func PreparePostConfig(config map[string]any, previousSave string) map[string]any {
	postConfig := make(map[string]any)

	setConfigMetadata(postConfig, previousSave)

	postConfig["config"] = prepareAPIConsoleConfig(config)

	setFastDataConfig(postConfig, config)

	setExtensionsAndMicrofrontendConfig(postConfig, config)

	return postConfig
}

func preparePostConfig(config map[string]any, previousSave string) map[string]any {
	return PreparePostConfig(config, previousSave)
}

func setConfigMetadata(postConfig map[string]any, previousSave string) {
	postConfig["title"] = "Import configurations"
	if previousSave != "" {
		postConfig["previousSave"] = previousSave
	}
}

func setFastDataConfig(postConfig map[string]any, config map[string]any) {
	if fastData, ok := config["fastDataConfig"]; ok && fastData != nil {
		postConfig["fastDataConfig"] = fastData
	}
}

func setExtensionsAndMicrofrontendConfig(postConfig map[string]any, config map[string]any) {
	postConfig["extensionsConfig"] = prepareExtensionsConfig(config)
	postConfig["microfrontendPluginsConfig"] = prepareMicrofrontendConfig(config)
}

func showConfirmationAndPrompt(opts *ConfigImportOptions, revision string, writer io.Writer) bool {
	fmt.Fprintln(writer, "=== CONFIGURATION IMPORT SUMMARY ===")
	fmt.Fprintln(writer, "The following configurations will be imported:")
	fmt.Fprintln(writer, "")

	configCount := displayConfigurationSummary(opts, writer)

	fmt.Fprintln(writer, "")
	fmt.Fprintf(writer, "Total configurations to import: %d\n", configCount)

	displayTargetInfo(revision, opts.Environment, writer)

	fmt.Fprintln(writer, "")
	fmt.Fprint(writer, "Do you want to proceed with this import? [y/N]: ")

	return getUserConfirmation()
}

func displayConfigurationSummary(opts *ConfigImportOptions, writer io.Writer) int {
	configCount := 0

	configsToDisplay := []struct {
		path string
		name string
	}{
		{opts.FlowManagerConfigPath, "Flow Manager Configuration"},
		{opts.RbacManagerConfigPath, "RBAC Manager Configuration"},
		{opts.BackofficeConfigPath, "Backoffice Configuration"},
		{opts.FastDataConfigPath, "Fast Data Configuration"},
		{opts.APIConsoleConfigPath, "API Console Configuration"},
	}

	for _, config := range configsToDisplay {
		if config.path != "" {
			fmt.Fprintf(writer, "  ✓ %s\n", config.name)
			fmt.Fprintf(writer, "    Source: %s\n", config.path)
			configCount++
		}
	}

	return configCount
}

func displayTargetInfo(revision, environment string, writer io.Writer) {
	if len(revision) > 0 {
		fmt.Fprintf(writer, "Target: Revision %s\n", revision)
	} else {
		fmt.Fprintf(writer, "Target: Environment %s\n", environment)
	}
}

func getUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
