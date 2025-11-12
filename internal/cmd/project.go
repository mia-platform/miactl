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
	"github.com/mia-platform/miactl/internal/cmd/project"
	"github.com/spf13/cobra"
)

func ProjectCmd(o *clioptions.CLIOptions) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Mia-Platform Console projects",
		Long: `Manage Mia-Platform Console projects.

Projects contains the configurations for the various Services, APIs, CronJobs and the other
resources that make up the applications of a specific company.
		`,
	}

	// add cmd flags
	flags := projectCmd.PersistentFlags()
	o.AddConnectionFlags(flags)
	o.AddContextFlags(flags)
	o.AddCompanyFlags(flags)

	// add sub commands
	projectCmd.AddCommand(
		project.ListCmd(o),
		project.IAMCmd(o),
		project.ImportCmd(o),
		project.DescribeCmd(o),
		project.ApplyCmd(o),
		project.VersionCmd(o),
	)

	return projectCmd
}
