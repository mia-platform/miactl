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
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewBasicLoginCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "basic [FLAGS]",
		Short: "set up M2M basic authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Context == "" {
				currentContext, err := context.GetCurrentContext()
				if err != nil {
					return err
				}
				options.Context = currentContext
			}
			home, err := homedir.Dir()
			if err != nil {
				return err
			}
			credentialsAbsPath := path.Join(home, M2MCredentialsPath)
			_, err = updateBasicCredentials(credentialsAbsPath, *options)
			if err != nil {
				return err
			}
			fmt.Println("New M2M basic auth credentials added!")
			return nil
		},
	}
	options.AddBasicAuthFlags(cmd)
	err := cobra.MarkFlagRequired(cmd.PersistentFlags(), "client-id")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cobra.MarkFlagRequired(cmd.PersistentFlags(), "client-secret")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return cmd
}

func updateBasicCredentials(credentialsPath string, opts clioptions.CLIOptions) (*M2MAuthInfo, error) {
	credentialsMap, err := ReadCredentials(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials from file: %w", err)
	}
	newBasicAuth := M2MAuthInfo{
		AuthType: "basic",
		BasicAuth: BasicAuthCredentials{
			ClientID:     opts.BasicClientID,
			ClientSecret: opts.BasicClientSecret,
		},
	}
	credentialsMap[opts.Context] = newBasicAuth
	credBytes, err := yaml.Marshal(credentialsMap)
	if err != nil {
		return nil, fmt.Errorf("error marshaling credentials: %w", err)
	}
	err = os.WriteFile(credentialsPath, credBytes, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error writing credentials file: %w", err)
	}
	return &newBasicAuth, nil
}
