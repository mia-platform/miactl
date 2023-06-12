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

package cmd

import (
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/spf13/cobra"
)

func ContextCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Modify cli context",
		Long: `Modify cli context. The file where the context will be saved and read will be selected
with this order:

 1. if the --config flag is used that file will be selected
 2. if the $MIACTL_CONFIG environment is set with a valid path the context will be read and write
back in that file
 3. if the $XDG_CONFIG_HOME environment the $XDG_CONFIG_HOME/miactl/config file is used
 4. lastly if nothing of the precedent rules apply the path $HOME/.config/miactl/config is used
`,
	}
	// add cmd flags

	// add sub commnads
	cmd.AddCommand(
		context.AuthCmd(options),
		context.SetCmd(options),
		context.UseCmd(options),
		context.ListCmd(options),
	)

	return cmd
}
