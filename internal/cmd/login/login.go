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

package login

import (
	"errors"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// nolint gosec
const M2MCredentialsPath = ".config/miactl/credentials"

var ErrMissingCredentials = errors.New("missing credentials for current and default context")

func NewLoginCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "set up authentication with M2M credentials (basic or jwt)",
	}

	options.AddContextFlags(cmd)
	cmd.AddCommand(NewBasicLoginCmd(options))

	return cmd
}

func ReadCredentials(credentialsPath string) (map[string]M2MAuthInfo, error) {
	yamlCredentials, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}
	var credentialsMap map[string]M2MAuthInfo
	err = yaml.Unmarshal(yamlCredentials, &credentialsMap)
	return credentialsMap, err
}

func GetCredentialsFromFile(credentialsPath, context string) (M2MAuthInfo, error) {
	credentialsMap, err := ReadCredentials(credentialsPath)
	if err != nil {
		return M2MAuthInfo{}, err
	}
	if credential, found := credentialsMap[context]; found {
		return credential, nil
	}
	if credential, found := credentialsMap["default"]; found {
		return credential, nil
	}
	return M2MAuthInfo{}, ErrMissingCredentials
}
