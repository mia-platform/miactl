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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportConfigCmdCreation(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, ImportConfigCmd(clioptions.NewCLIOptions()))
}

func TestImportConfigValidation(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "", "", &ConfigImportOptions{}, nil), "missing project id")
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "projectID", "", &ConfigImportOptions{}, nil), "either revision or environment must be specified")
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "projectID", "revision", &ConfigImportOptions{Environment: "env"}, nil), "either revision or environment must be specified")
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "projectID", "revision", &ConfigImportOptions{}, nil), "missing configuration files")

	opts := &ConfigImportOptions{
		FlowManagerConfigPath: "non-existent-file.json",
	}
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "projectID", "revision", opts, nil), "flow manager config file not found")

	opts = &ConfigImportOptions{
		FlowManagerConfigPath: "non-existent-file.json",
		Environment:           "env",
	}
	assert.ErrorContains(t, importConfigFromFiles(ctx, nil, "projectID", "", opts, nil), "flow manager config file not found")
}

func TestFileValidation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid-config.json")
	err := os.WriteFile(validFile, []byte("{}"), 0644)
	require.NoError(t, err)

	opts := &ConfigImportOptions{
		FlowManagerConfigPath: validFile,
	}
	err = validateFilePaths(opts)
	assert.NoError(t, err)

	opts = &ConfigImportOptions{
		FlowManagerConfigPath: filepath.Join(tempDir, "non-existent.json"),
	}
	err = validateFilePaths(opts)
	assert.Error(t, err)
}

func TestReadJSONFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	validJSON := filepath.Join(tempDir, "valid.json")
	err := os.WriteFile(validJSON, []byte(`{"key": "value"}`), 0644)
	require.NoError(t, err)

	invalidExt := filepath.Join(tempDir, "invalid.txt")
	err = os.WriteFile(invalidExt, []byte(`{"key": "value"}`), 0644)
	require.NoError(t, err)

	var result map[string]interface{}
	err = readJSONFile(validJSON, &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])

	err = readJSONFile(invalidExt, &result)
	assert.ErrorContains(t, err, "file must be .json")
}

func TestPreparePostConfig(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		"committedDate":    "2023-01-01",
		"lastCommitAuthor": "Test User",
		"platformVersion":  "v1.0.0",
		"commitId":         "12345",
		"apiField1":        "value1",
		"apiField2":        "value2",
		"fastDataConfig": map[string]any{
			"fastDataField1": "fdValue1",
		},
		"extensionsConfig": map[string]any{
			"extensionField1": "extValue1",
		},
		"microfrontendPluginsConfig": map[string]any{
			"existingPlugin": "pluginValue",
		},
	}

	postConfig := preparePostConfig(config, "")

	assert.Equal(t, "Import configurations", postConfig["title"])

	assert.Equal(t, "value1", postConfig["config"].(map[string]any)["apiField1"])
	assert.Equal(t, "value2", postConfig["config"].(map[string]any)["apiField2"])
	assert.Equal(t, "12345", postConfig["config"].(map[string]any)["commitId"])

	assert.Equal(t, "fdValue1", postConfig["fastDataConfig"].(map[string]any)["fastDataField1"])
	assert.Equal(t, "extValue1", postConfig["extensionsConfig"].(map[string]any)["extensionField1"])
	assert.Equal(t, "pluginValue", postConfig["microfrontendPluginsConfig"].(map[string]any)["existingPlugin"])

	_, hasCommittedDate := postConfig["config"].(map[string]any)["committedDate"]
	_, hasLastCommitAuthor := postConfig["config"].(map[string]any)["lastCommitAuthor"]
	_, hasPlatformVersion := postConfig["config"].(map[string]any)["platformVersion"]
	assert.False(t, hasCommittedDate)
	assert.False(t, hasLastCommitAuthor)
	assert.False(t, hasPlatformVersion)
}

func TestAPIConsoleConfigMerging(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	apiConsoleFile := filepath.Join(tempDir, "apiConsole.json")
	err := os.WriteFile(apiConsoleFile, []byte(`{"consoleField1": "consoleValue1", "consoleField2": "consoleValue2"}`), 0644)
	require.NoError(t, err)

	config := map[string]any{
		"apiField1": "value1",
		"apiField2": "value2",
	}

	var apiConsoleData map[string]any
	err = readJSONFile(apiConsoleFile, &apiConsoleData)
	require.NoError(t, err)

	// Simulate merging APIConsoleConfig into the main config
	for key, value := range apiConsoleData {
		config[key] = value
	}

	postConfig := preparePostConfig(config, "")

	assert.Equal(t, "value1", postConfig["config"].(map[string]any)["apiField1"])
	assert.Equal(t, "value2", postConfig["config"].(map[string]any)["apiField2"])
	assert.Equal(t, "consoleValue1", postConfig["config"].(map[string]any)["consoleField1"])
	assert.Equal(t, "consoleValue2", postConfig["config"].(map[string]any)["consoleField2"])
}

func TestEndpointSelection(t *testing.T) {
	t.Parallel()

	projectID := "project1"
	revision := "rev1"
	environment := ""
	endpoint := ""

	if len(revision) > 0 {
		endpoint = fmt.Sprintf(configRevisionEndpointTemplate, projectID, revision)
	} else {
		endpoint = fmt.Sprintf(configEnvironmentEndpointTemplate, projectID, environment)
	}

	assert.Equal(t, "/api/backend/projects/project1/revisions/rev1/configuration", endpoint)

	revision = ""
	environment = "env1"

	if len(revision) > 0 {
		endpoint = fmt.Sprintf(configRevisionEndpointTemplate, projectID, revision)
	} else {
		endpoint = fmt.Sprintf(configEnvironmentEndpointTemplate, projectID, environment)
	}

	assert.Equal(t, "/api/projects/project1/environments/env1/configuration", endpoint)
}

func TestConfirmationPrompt(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	flowManagerFile := filepath.Join(tempDir, "flowManager.json")
	err := os.WriteFile(flowManagerFile, []byte(`{"flowSetting": "value1"}`), 0644)
	require.NoError(t, err)

	rbacManagerFile := filepath.Join(tempDir, "rbacManager.json")
	err = os.WriteFile(rbacManagerFile, []byte(`{"rbacSetting": "value2"}`), 0644)
	require.NoError(t, err)

	opts := &ConfigImportOptions{
		FlowManagerConfigPath: flowManagerFile,
		RbacManagerConfigPath: rbacManagerFile,
		SkipConfirmation:      true,
		Environment:           "test-env",
	}

	err = validateFilePaths(opts)
	assert.NoError(t, err)

	assert.True(t, opts.SkipConfirmation)

	opts.SkipConfirmation = false
	assert.False(t, opts.SkipConfirmation)
}

func TestValidateFilePaths(t *testing.T) {
	t.Parallel()

	opts := &ConfigImportOptions{
		FlowManagerConfigPath: "non-existent-file.json",
	}
	err := validateFilePaths(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flow manager config file not found")

	opts = &ConfigImportOptions{}
	err = validateFilePaths(opts)
	assert.NoError(t, err)
}

func TestConfirmationFlow(t *testing.T) {
	t.Parallel()

	flowManagerFile := filepath.Join("..", "..", "..", "examples", "configuration-flow-manager.json")
	rbacManagerFile := filepath.Join("..", "..", "..", "examples", "configuration-rbac-manager.json")
	fastDataFile := filepath.Join("..", "..", "..", "examples", "configuration-fast-data.json")
	backofficeFile := filepath.Join("..", "..", "..", "examples", "configuration-backoffice.json")

	require.FileExists(t, flowManagerFile)
	require.FileExists(t, rbacManagerFile)
	require.FileExists(t, fastDataFile)
	require.FileExists(t, backofficeFile)

	opts := &ConfigImportOptions{
		FlowManagerConfigPath: flowManagerFile,
		RbacManagerConfigPath: rbacManagerFile,
		FastDataConfigPath:    fastDataFile,
		BackofficeConfigPath:  backofficeFile,
		SkipConfirmation:      true,
		Environment:           "test-env",
	}

	err := validateFilePaths(opts)
	assert.NoError(t, err)

	var flowManagerData map[string]any
	err = readJSONFile(flowManagerFile, &flowManagerData)
	assert.NoError(t, err)
	assert.NotNil(t, flowManagerData["config"])
	flowConfig := flowManagerData["config"].(map[string]any)
	assert.Equal(t, "2.3.1", flowConfig["version"])

	var rbacManagerData map[string]any
	err = readJSONFile(rbacManagerFile, &rbacManagerData)
	assert.NoError(t, err)
	assert.NotNil(t, rbacManagerData["config"])
	rbacConfig := rbacManagerData["config"].(map[string]any)
	assert.NotNil(t, rbacConfig["permissions"])

	var fastDataData map[string]any
	err = readJSONFile(fastDataFile, &fastDataData)
	assert.NoError(t, err)
	assert.NotNil(t, fastDataData["config"])
	fastDataConfig := fastDataData["config"].(map[string]any)
	assert.Equal(t, "2.2.0", fastDataConfig["version"])

	var backofficeData map[string]any
	err = readJSONFile(backofficeFile, &backofficeData)
	assert.NoError(t, err)
	assert.NotNil(t, backofficeData["config"])
	backofficeConfig := backofficeData["config"].(map[string]any)
	assert.Equal(t, "1.8.0", backofficeConfig["version"])

	projectConfig := map[string]any{
		"commitId":                   "test-commit-id",
		"version":                    "1.0.0",
		"microfrontendPluginsConfig": make(map[string]any),
	}

	microfrontendConfig := projectConfig["microfrontendPluginsConfig"].(map[string]any)
	microfrontendConfig["flowManagerConfig"] = flowConfig
	microfrontendConfig["rbacManagerConfig"] = rbacConfig
	microfrontendConfig["backofficeConfigurations"] = backofficeConfig
	projectConfig["fastDataConfig"] = fastDataConfig

	postConfig := preparePostConfig(projectConfig, "test-previous-save")

	assert.Equal(t, "Import configurations", postConfig["title"])
	assert.Equal(t, "test-previous-save", postConfig["previousSave"])
	assert.NotNil(t, postConfig["config"])
	assert.NotNil(t, postConfig["fastDataConfig"])
	assert.NotNil(t, postConfig["microfrontendPluginsConfig"])
	assert.NotNil(t, postConfig["extensionsConfig"])

	mfeConfig := postConfig["microfrontendPluginsConfig"].(map[string]any)
	assert.NotNil(t, mfeConfig["flowManagerConfig"])
	assert.NotNil(t, mfeConfig["rbacManagerConfig"])
	assert.NotNil(t, mfeConfig["backofficeConfigurations"])

	fastDataInPayload := postConfig["fastDataConfig"].(map[string]any)
	assert.Equal(t, "2.2.0", fastDataInPayload["version"])

	mainConfig := postConfig["config"].(map[string]any)
	_, hasFastData := mainConfig["fastDataConfig"]
	_, hasMfe := mainConfig["microfrontendPluginsConfig"]
	_, hasExtensions := mainConfig["extensionsConfig"]
	assert.False(t, hasFastData)
	assert.False(t, hasMfe)
	assert.False(t, hasExtensions)

	assert.Equal(t, "test-commit-id", mainConfig["commitId"])
	assert.Equal(t, "1.0.0", mainConfig["version"])
}
