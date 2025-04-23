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

package util

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"

	"github.com/spf13/cobra"
)

func ShowDeprecatedMessage(opts *clioptions.CLIOptions) func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		restConfig, err := opts.ToRESTConfig()
		cobra.CheckErr(err)
		client, err := client.APIClientForConfig(restConfig)
		cobra.CheckErr(err)

		canUseNewApi, error := VersionCheck(cmd.Context(), client, 14, 0)
		if error == nil && canUseNewApi {
			writer := cmd.ErrOrStderr()
			fmt.Fprint(writer, "\nThe command you are using is deprecated. Please use 'miactl catalog' instead.")
		}
	}
}
