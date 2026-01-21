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

package deploy

import (
	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/clioptions"
)

func NewDeployCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Manage project deployments",
		Long: `Manage project deployments.

Can trigger deployments to specific environments and monitor their status.`,
	}

	cmd.AddCommand(
		triggerCmd(options),
		newStatusAddCmd(options),
		latestDeploymentCmd(options),
	)

	return cmd
}
