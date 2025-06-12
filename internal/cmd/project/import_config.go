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

// ConfigImportOptions contains the file paths for the various configuration files
type ConfigImportOptions struct {
	FlowManagerConfigPath string
	RbacManagerConfigPath string
	BackofficeConfigPath  string
	FastDataConfigPath    string
	APIConsoleConfigPath  string
	Environment           string // Environment ID for saving to environment configuration
	SkipConfirmation      bool   // If true, skip interactive confirmation
}

// ImportConfigCmd return a cobra command for importing configurations from JSON files
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

	// add cmd flags
	flags := cmd.Flags()
	o.AddProjectFlags(flags)
	flags.StringVar(&o.Revision, "revision", "", "revision of the commit where to import the configuration (mutually exclusive with --environment)")
	flags.StringVar(&importOptions.Environment, "environment", "", "environment where to import the configuration (mutually exclusive with --revision)")

	// add config file flags
	flags.StringVar(&importOptions.FlowManagerConfigPath, "flow-manager-config", "", "path to flowManagerConfig.json file")
	flags.StringVar(&importOptions.RbacManagerConfigPath, "rbac-manager-config", "", "path to rbacManagerConfig.json file")
	flags.StringVar(&importOptions.BackofficeConfigPath, "backoffice-config", "", "path to backofficeConfigurations.json file")
	flags.StringVar(&importOptions.FastDataConfigPath, "fast-data-config", "", "path to fast-data-config.json file")
	flags.StringVar(&importOptions.APIConsoleConfigPath, "api-console-config", "", "path to api-console-config.json file")

	// add confirmation flag
	flags.BoolVarP(&importOptions.SkipConfirmation, "yes", "y", false, "skip interactive confirmation and proceed with import")

	return cmd
}

func importConfigFromFiles(ctx context.Context, client *client.APIClient, projectID, revision string, opts *ConfigImportOptions, writer io.Writer) error {
	if len(projectID) == 0 {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	if (len(revision) == 0 && len(opts.Environment) == 0) || (len(revision) > 0 && len(opts.Environment) > 0) {
		return fmt.Errorf("either revision or environment must be specified, but not both")
	}

	if opts.FlowManagerConfigPath == "" &&
		opts.RbacManagerConfigPath == "" &&
		opts.BackofficeConfigPath == "" &&
		opts.FastDataConfigPath == "" &&
		opts.APIConsoleConfigPath == "" {
		return fmt.Errorf("missing configuration files, please specify at least one configuration file")
	}

	if err := validateFilePaths(opts); err != nil {
		return err
	}

	var endpoint string
	if len(revision) > 0 {
		endpoint = fmt.Sprintf(configRevisionEndpointTemplate, projectID, revision)
	} else {
		endpoint = fmt.Sprintf(configEnvironmentEndpointTemplate, projectID, opts.Environment)
	}

	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)

	if err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	projectConfig := make(map[string]any)
	if err := response.ParseResponse(&projectConfig); err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	var previousSave string
	if changesId, ok := projectConfig["changesId"].(string); ok {
		previousSave = changesId
	}

	if projectConfig["microfrontendPluginsConfig"] == nil {
		projectConfig["microfrontendPluginsConfig"] = make(map[string]any)
	}

	microfrontendConfig, _ := projectConfig["microfrontendPluginsConfig"].(map[string]any)

	if opts.FlowManagerConfigPath != "" {
		var flowManagerData map[string]any
		if err := readJSONFile(opts.FlowManagerConfigPath, &flowManagerData); err != nil {
			return fmt.Errorf("error reading flow manager config: %w", err)
		}
		if config, ok := flowManagerData["config"]; ok {
			microfrontendConfig["flowManagerConfig"] = config
		} else {
			microfrontendConfig["flowManagerConfig"] = flowManagerData
		}
	}

	if opts.RbacManagerConfigPath != "" {
		var rbacManagerData map[string]any
		if err := readJSONFile(opts.RbacManagerConfigPath, &rbacManagerData); err != nil {
			return fmt.Errorf("error reading RBAC manager config: %w", err)
		}
		if config, ok := rbacManagerData["config"]; ok {
			microfrontendConfig["rbacManagerConfig"] = config
		} else {
			microfrontendConfig["rbacManagerConfig"] = rbacManagerData
		}
	}

	if opts.BackofficeConfigPath != "" {
		var backofficeData map[string]any
		if err := readJSONFile(opts.BackofficeConfigPath, &backofficeData); err != nil {
			return fmt.Errorf("error reading backoffice configurations: %w", err)
		}
		if config, ok := backofficeData["config"]; ok {
			microfrontendConfig["backofficeConfigurations"] = config
		} else {
			microfrontendConfig["backofficeConfigurations"] = backofficeData
		}
	}

	if opts.FastDataConfigPath != "" {
		var fastDataData map[string]any
		if err := readJSONFile(opts.FastDataConfigPath, &fastDataData); err != nil {
			return fmt.Errorf("error reading fast data config: %w", err)
		}
		if config, ok := fastDataData["config"]; ok {
			projectConfig["fastDataConfig"] = config
		} else {
			projectConfig["fastDataConfig"] = fastDataData
		}
	}

	if opts.APIConsoleConfigPath != "" {
		var apiConsoleData map[string]any
		if err := readJSONFile(opts.APIConsoleConfigPath, &apiConsoleData); err != nil {
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

	if !opts.SkipConfirmation {
		if !showConfirmationAndPrompt(opts, revision, writer) {
			fmt.Fprintln(writer, "Operation cancelled by user")
			return nil
		}
	}

	postConfig := preparePostConfig(projectConfig, previousSave)

	body, err := resources.EncodeResourceToJSON(postConfig)
	if err != nil {
		return fmt.Errorf("cannot encode project configuration: %w", err)
	}

	response, err = client.
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

func validateFilePaths(opts *ConfigImportOptions) error {
	if opts.FlowManagerConfigPath != "" {
		if _, err := os.Stat(opts.FlowManagerConfigPath); err != nil {
			return fmt.Errorf("flow manager config file not found: %w", err)
		}
	}
	if opts.RbacManagerConfigPath != "" {
		if _, err := os.Stat(opts.RbacManagerConfigPath); err != nil {
			return fmt.Errorf("RBAC manager config file not found: %w", err)
		}
	}
	if opts.BackofficeConfigPath != "" {
		if _, err := os.Stat(opts.BackofficeConfigPath); err != nil {
			return fmt.Errorf("backoffice configurations file not found: %w", err)
		}
	}
	if opts.FastDataConfigPath != "" {
		if _, err := os.Stat(opts.FastDataConfigPath); err != nil {
			return fmt.Errorf("fast data config file not found: %w", err)
		}
	}
	if opts.APIConsoleConfigPath != "" {
		if _, err := os.Stat(opts.APIConsoleConfigPath); err != nil {
			return fmt.Errorf("API console config file not found: %w", err)
		}
	}
	return nil
}

func readJSONFile(filePath string, target interface{}) error {
	ext := filepath.Ext(filePath)
	if ext != ".json" {
		return fmt.Errorf("file must be .json: %s", filePath)
	}

	// Use json.Unmarshal directly to ensure consistent number handling
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
	}

	return nil
}

func preparePostConfig(config map[string]any, previousSave string) map[string]any {
	postConfig := make(map[string]any)

	postConfig["title"] = "Import configurations"
	if previousSave != "" {
		postConfig["previousSave"] = previousSave
	}

	apiConsoleConfig := make(map[string]any)

	for k, v := range config {
		if k != "fastDataConfig" && k != "microfrontendPluginsConfig" && k != "extensionsConfig" &&
			k != "committedDate" && k != "lastCommitAuthor" && k != "platformVersion" && k != "changesId" {
			apiConsoleConfig[k] = v
		}
	}

	postConfig["config"] = apiConsoleConfig

	if fastData, ok := config["fastDataConfig"]; ok && fastData != nil {
		postConfig["fastDataConfig"] = fastData
	}

	if extensions, ok := config["extensionsConfig"]; ok && extensions != nil {
		postConfig["extensionsConfig"] = extensions
	} else {
		postConfig["extensionsConfig"] = map[string]any{
			"files": make(map[string]any),
		}
	}

	mfeConfig := make(map[string]any)
	if mfe, ok := config["microfrontendPluginsConfig"]; ok && mfe != nil {
		mfeConfig = mfe.(map[string]any)
	}

	postConfig["microfrontendPluginsConfig"] = mfeConfig

	return postConfig
}

func showConfirmationAndPrompt(opts *ConfigImportOptions, revision string, writer io.Writer) bool {
	fmt.Fprintln(writer, "=== CONFIGURATION IMPORT SUMMARY ===")
	fmt.Fprintln(writer, "The following configurations will be imported:")
	fmt.Fprintln(writer, "")

	configCount := 0

	if opts.FlowManagerConfigPath != "" {
		fmt.Fprintf(writer, "  ✓ Flow Manager Configuration\n")
		fmt.Fprintf(writer, "    Source: %s\n", opts.FlowManagerConfigPath)
		configCount++
	}

	if opts.RbacManagerConfigPath != "" {
		fmt.Fprintf(writer, "  ✓ RBAC Manager Configuration\n")
		fmt.Fprintf(writer, "    Source: %s\n", opts.RbacManagerConfigPath)
		configCount++
	}

	if opts.BackofficeConfigPath != "" {
		fmt.Fprintf(writer, "  ✓ Backoffice Configuration\n")
		fmt.Fprintf(writer, "    Source: %s\n", opts.BackofficeConfigPath)
		configCount++
	}

	if opts.FastDataConfigPath != "" {
		fmt.Fprintf(writer, "  ✓ Fast Data Configuration\n")
		fmt.Fprintf(writer, "    Source: %s\n", opts.FastDataConfigPath)
		configCount++
	}

	if opts.APIConsoleConfigPath != "" {
		fmt.Fprintf(writer, "  ✓ API Console Configuration\n")
		fmt.Fprintf(writer, "    Source: %s\n", opts.APIConsoleConfigPath)
		configCount++
	}

	fmt.Fprintln(writer, "")
	fmt.Fprintf(writer, "Total configurations to import: %d\n", configCount)

	if len(revision) > 0 {
		fmt.Fprintf(writer, "Target: Revision %s\n", revision)
	} else {
		fmt.Fprintf(writer, "Target: Environment %s\n", opts.Environment)
	}

	fmt.Fprintln(writer, "")
	fmt.Fprint(writer, "Do you want to proceed with this import? [y/N]: ")

	return getUserConfirmation()
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
