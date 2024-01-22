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

package iam

import (
	"context"
	"fmt"
	"os"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all IAM entity for a company",
		Long: `A Company can have associated different entities for managing the roles, this command will list
all of them noting the type and the current role associated with them`,

		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			entityTypes := map[string]bool{
				iam.GroupsEntityName:          options.ShowGroups,
				iam.UsersEntityName:           options.ShowUsers,
				iam.ServiceAccountsEntityName: options.ShowServiceAccounts,
			}

			return listAllIAMEntities(cmd.Context(), client, restConfig.CompanyID, restConfig.ProjectID, entityTypes)
		},
	}

	options.AddIAMListFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("users", "groups", "serviceAccounts")

	return cmd
}

func listAllIAMEntities(ctx context.Context, client *client.APIClient, companyID, projectID string, entityTypes map[string]bool) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}

	if len(projectID) == 0 {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	resp, err := iam.ListAllIAMEntities(ctx, client, companyID, []string{projectID}, entityTypes)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	rows, err := util.RowsForResources(resp, iam.RowForProjectIAMIdentity(projectID))
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetAutoWrapText(false)
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"ID", "Type", "Name", "Roles"})
	table.AppendBulk(rows)
	table.Render()
	return nil
}
