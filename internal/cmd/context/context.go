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

package context

import (
	"fmt"
	"strings"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewContextCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "perform operations on cluster contexts",
	}

	cmd.AddCommand(NewSetContextCmd(options))
	cmd.AddCommand(NewUseContextCmd(options))
	cmd.AddCommand(NewListContextsCmd(options))

	return cmd
}

func GetContextBaseURL(contextName string) (string, error) {
	context, err := contextLookUp(contextName)
	if err != nil {
		return "", fmt.Errorf("error while searching context in config file: %w", err)
	}
	return fmt.Sprint(context["endpoint"]), nil
}

func GetContextCompanyID(contextName string) (string, error) {
	context, err := contextLookUp(contextName)
	if err != nil {
		return "", fmt.Errorf("error while searching context in config file: %w", err)
	}
	if context["companyid"] == nil {
		return "", fmt.Errorf("please set a company ID for context %s", contextName)
	}
	return fmt.Sprint(context["companyid"]), nil
}

func GetCurrentContext() (string, error) {
	if viper.Get("current-context") == nil {
		return "", fmt.Errorf("current context is unset")
	}
	return fmt.Sprint(viper.Get("current-context")), nil
}

func GetContextProjectID(contextName string) (string, error) {
	context, err := contextLookUp(contextName)
	if err != nil {
		return "", fmt.Errorf("error while searching context in config file: %w", err)
	}
	if context["projectid"] == nil {
		return "", fmt.Errorf("please set a project ID for context %s", contextName)
	}
	return fmt.Sprint(context["projectid"]), nil
}

func getContextMap() (map[string]interface{}, error) {
	if viper.Get("contexts") == nil {
		return nil, fmt.Errorf("no context specified in config file")
	}
	return viper.Get("contexts").(map[string]interface{}), nil
}

// SetContextValues checks if a flag was explicitly set by the user, and loads the value
// from the config file otherwise
func SetContextValues(cmd *cobra.Command, currentContext string) error {
	var cValues = []string{"project-id", "company-id", "endpoint", "ca-cert", "insecure"}

	for _, val := range cValues {
		flag := cmd.Flag(val)
		if flag == nil {
			continue
		}
		viperKey := strings.ReplaceAll(val, "-", "")
		viperPath := fmt.Sprintf("contexts.%s.%s", currentContext, viperKey)
		if !flag.Changed && viper.IsSet(viperPath) {
			viperValue := viper.GetString(viperPath)
			if err := flag.Value.Set(viperValue); err != nil {
				return err
			}
		}
	}
	return nil
}
