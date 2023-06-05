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
	"os"
	"path/filepath"
)

const (
	configFileName   = "config"
	miactlFolderName = "miactl"
	credentials      = "credentials"
)

func homeFolderPath() string {
	home, err := os.UserHomeDir()
	if home == "" || err != nil {
		home = "/"
	}

	return home
}

func configFolderPath() string {
	configFolderPath := os.Getenv("XDG_CONFIG_HOME")
	if configFolderPath == "" {
		configFolderPath = filepath.Join(homeFolderPath(), ".config")
	}
	return configFolderPath
}

func cacheFolderPath() string {
	cacheFolderPath := os.Getenv("XDG_CACHE_HOME")
	if cacheFolderPath == "" {
		cacheFolderPath = filepath.Join(homeFolderPath(), ".cache")
	}
	return cacheFolderPath
}

func ConfigFilePath() string {
	return filepath.Join(configFolderPath(), miactlFolderName, configFileName)
}

func ConfigFilePathString() string {
	return filepath.Join("$HOME", miactlFolderName, configFileName)
}

func CredentialsFilePath() string {
	return filepath.Join(configFolderPath(), miactlFolderName, credentials)
}

func CacheFolderPath() string {
	return filepath.Join(cacheFolderPath(), miactlFolderName)
}
