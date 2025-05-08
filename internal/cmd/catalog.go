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
	"github.com/mia-platform/miactl/internal/cmd/catalog"
	catalog_apply "github.com/mia-platform/miactl/internal/cmd/catalog/apply"
	"github.com/spf13/cobra"
)

func CatalogCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "View and manage Catalog items. This command is available from Mia-Platform Console v14.0.0.",
	}

	// add cmd flags
	flags := cmd.PersistentFlags()
	options.AddConnectionFlags(flags)
	options.AddCompanyFlags(flags)
	options.AddContextFlags(flags)

	// add sub commands
	cmd.AddCommand(catalog.ListCmd(options))
	cmd.AddCommand(catalog.GetCmd(options))
	cmd.AddCommand(catalog.DeleteCmd(options))
	cmd.AddCommand(catalog_apply.ApplyCmd(options))
	cmd.AddCommand(catalog.ListVersionCmd(options))

	return cmd
}
