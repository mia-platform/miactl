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

package cliconfig

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/cliconfig/api"
	"gopkg.in/yaml.v3"
)

const ConfigPathEnvVarName = "MIACONFIG"

type ConfigPathLocator struct {
	ExplicitPath string

	filePath string
}

func NewConfigPathLocator() *ConfigPathLocator {
	filePath := ConfigFilePath()

	if envVarFiles := os.Getenv(ConfigPathEnvVarName); len(envVarFiles) != 0 {
		filePath = filepath.SplitList(envVarFiles)[0]
	}

	return &ConfigPathLocator{
		filePath: filePath,
	}
}

func (cr *ConfigPathLocator) DefaultConfigPath() string {
	return ConfigFilePathString()
}

func (cr *ConfigPathLocator) ReadConfig() (*api.Config, error) {
	return readFile(cr.ConfigLocation())
}

func (cr *ConfigPathLocator) WriteConfig(config *api.Config) error {
	return writeFile(cr.ConfigLocation(), config)
}

func (cr *ConfigPathLocator) ConfigLocation() string {
	if len(cr.ExplicitPath) > 0 {
		return cr.ExplicitPath
	}

	return cr.filePath
}

func readFile(path string) (*api.Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return api.NewConfig(), nil
		default:
			return nil, err
		}
	}

	if len(configData) == 0 {
		return api.NewConfig(), nil
	}

	config := new(api.Config)
	decoder := yaml.NewDecoder(bytes.NewBuffer(configData))
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func writeFile(path string, config *api.Config) error {
	// be sure that the path to file actually exists
	// if not try to create it
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	configBuffer := bytes.NewBuffer([]byte{})
	encoder := yaml.NewEncoder(configBuffer)

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return os.WriteFile(path, configBuffer.Bytes(), 0600)
}
