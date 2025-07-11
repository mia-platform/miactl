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

package configuration

import "fmt"

const (
	ErrNotValidConfiguration = "provided configuration is not valid"
)

type Configuration struct {
	Config                     map[string]any `json:"config" yaml:"config"`
	FastDataConfig             map[string]any `json:"fastDataConfig,omitempty" yaml:"fastDataConfig,omitempty"`
	MicrofrontendPluginsConfig map[string]any `json:"microfrontendPluginsConfig,omitempty" yaml:"microfrontendPluginsConfig,omitempty"`
	ExtensionsConfig           map[string]any `json:"extensionsConfig,omitempty" yaml:"extensionsConfig,omitempty"`
}

// BuildDescribeConfiguration builds a DescribeConfiguration from a configuration
func BuildDescribeConfiguration(rawConfig map[string]any) (*Configuration, error) {
	return describeConfigurationAdapter(rawConfig)
}

func BuildDescribeFromFlatConfiguration(rawConfig map[string]any) (*Configuration, error) {
	return flatConfigurationAdapter(rawConfig)
}

// describeConfigurationAdapter adapts an already describe configuration map to the expected DescribeConfiguration type.
// This is used to adapt output from the describe command (e.g., when reading the describe output from a file) to a valid DescribeConfiguration structure.
func describeConfigurationAdapter(config map[string]any) (*Configuration, error) {
	applyConfig := &Configuration{}

	configSection, hasConfigKey := config["config"]
	if !hasConfigKey {
		return nil, fmt.Errorf("%s: %s", ErrNotValidConfiguration, "'config' key not found")
	}

	configSectionMap, ok := configSection.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: %s", ErrNotValidConfiguration, "'config' key is not a valid map[string]any")
	}

	baseConfig := make(map[string]any)
	for key, value := range configSectionMap {
		baseConfig[key] = value
	}

	if fastDataConfig, ok := getConfig("fastDataConfig", config); ok && fastDataConfig != nil {
		applyConfig.FastDataConfig = fastDataConfig
	}

	if extensionsConfig, ok := getConfig("extensionsConfig", config); ok && extensionsConfig != nil {
		applyConfig.ExtensionsConfig = extensionsConfig
	}

	if microfrontendPluginsConfig, ok := getConfig("microfrontendPluginsConfig", config); ok && microfrontendPluginsConfig != nil {
		applyConfig.MicrofrontendPluginsConfig = microfrontendPluginsConfig
	}

	applyConfig.Config = baseConfig
	return applyConfig, nil
}

// AdaptFlatConfiguration adapts a flat configuration map to the structured format.
// This is used to map the GET /configuration response into the DescribeConfiguration structure.
func flatConfigurationAdapter(config map[string]any) (*Configuration, error) {
	applyConfig := &Configuration{}

	baseConfig := make(map[string]any)
	for key, value := range config {
		baseConfig[key] = value
	}

	if fastDataConfig, ok := getConfig("fastDataConfig", baseConfig); ok && fastDataConfig != nil {
		applyConfig.FastDataConfig = fastDataConfig
		delete(baseConfig, "fastDataConfig")
	}

	if extensionsConfig, ok := getConfig("extensionsConfig", baseConfig); ok && extensionsConfig != nil {
		applyConfig.ExtensionsConfig = extensionsConfig
		delete(baseConfig, "extensionsConfig")
	}

	if microfrontendPluginsConfig, ok := getConfig("microfrontendPluginsConfig", baseConfig); ok && microfrontendPluginsConfig != nil {
		applyConfig.MicrofrontendPluginsConfig = microfrontendPluginsConfig
		delete(baseConfig, "microfrontendPluginsConfig")
	}

	applyConfig.Config = baseConfig
	return applyConfig, nil
}

func getConfig(configKey string, config map[string]any) (map[string]any, bool) {
	if configValue, ok := config[configKey]; ok {
		if configMap, ok := configValue.(map[string]any); ok {
			return configMap, true
		}
	}

	return nil, false
}
