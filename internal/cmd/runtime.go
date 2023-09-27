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
	"github.com/mia-platform/miactl/internal/cmd/cronjobs"
	"github.com/mia-platform/miactl/internal/cmd/environments"
	"github.com/mia-platform/miactl/internal/cmd/jobs"
	"github.com/mia-platform/miactl/internal/cmd/pods"
	"github.com/spf13/cobra"
)

func RuntimeCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Manage resources deployed with Mia-Platform Console",
		Long: `Manage resources deployed with Mia-Platform Console.

Once a project from Mia-Platform Console is deployed at least once, you can manage and monitor
the resources generated, like Pods, Cronjobs and logs.
`,
	}

	// add cmd flags
	flags := cmd.PersistentFlags()
	o.AddConnectionFlags(flags)
	o.AddContextFlags(flags)
	o.AddCompanyFlags(flags)
	o.AddProjectFlags(flags)

	// add sub commands
	cmd.AddCommand(
		environments.EnvironmentCmd(o),
		pods.Command(o),
		cronjobs.Command(o),
		jobs.Command(o),
	)

	return cmd
}
