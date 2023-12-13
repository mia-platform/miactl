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
	"github.com/mia-platform/miactl/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listAllIAMEntitiesTemplate = "/api/companies/%s/identities"
	listUsersEntityTemplate    = "/api/companies/%s/users"
	listGroupsEntityTemplate   = "/api/companies/%s/groups"

	GroupsEntityName          = "group"
	UsersEntityName           = "user"
	ServiceAccountsEntityName = "serviceAccount"
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
				GroupsEntityName:          options.ShowGroups,
				UsersEntityName:           options.ShowUsers,
				ServiceAccountsEntityName: options.ShowServiceAccounts,
			}

			return listAllIAMEntities(cmd.Context(), client, restConfig.CompanyID, entityTypes)
		},
	}

	options.AddIAMListFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("users", "groups", "serviceAccounts")

	cmd.AddCommand(
		listUsersCmd(options),
		listGroupsCmd(options),
	)

	return cmd
}

func listUsersCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "List all users that have access to the company, directly or via a group",
		Long:  "List all users that have access to the company, directly or via a group",

		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			return listSpecificEntities(cmd.Context(), client, restConfig.CompanyID, UsersEntityName)
		},
	}

	return cmd
}

func listGroupsCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "List all groups that have access to the company",
		Long:  "List all groups that have access to the company",

		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			return listSpecificEntities(cmd.Context(), client, restConfig.CompanyID, GroupsEntityName)
		},
	}

	return cmd
}

func listAllIAMEntities(ctx context.Context, client *client.APIClient, companyID string, entityTypes map[string]bool) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}

	request := client.
		Get().
		APIPath(fmt.Sprintf(listAllIAMEntitiesTemplate, companyID))

	for entityName, enabled := range entityTypes {
		if !enabled {
			continue
		}
		request.SetParam("identityType", entityName)
	}

	resp, err := request.Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	rows, err := util.RowsForResources(resp, rowForIAMIdentity)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Type", "Name", "Roles"})
	table.AppendBulk(rows)
	table.Render()
	return nil
}

func listSpecificEntities(ctx context.Context, client *client.APIClient, companyID string, entityType string) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}

	var apiPathTemplate string

	switch entityType {
	case UsersEntityName:
		apiPathTemplate = listUsersEntityTemplate
	case GroupsEntityName:
		apiPathTemplate = listGroupsEntityTemplate
	default:
		return fmt.Errorf("unknown IAM entity")
	}

	response, err := client.
		Get().
		APIPath(fmt.Sprintf(apiPathTemplate, companyID)).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := response.Error(); err != nil {
		return err
	}

	var tableHeaders []string
	var rows [][]string
	switch entityType {
	case UsersEntityName:
		tableHeaders = []string{"Name", "Email", "Roles", "Groups", "Last Login"}
		rows, err = util.RowsForResources(response, rowForUserIdentity)
	case GroupsEntityName:
		tableHeaders = []string{"Name", "Role", "Members"}
		rows, err = util.RowsForResources(response, rowForGroupIdentity)
	}

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader(tableHeaders)
	table.SetAutoWrapText(false)
	table.AppendBulk(rows)
	table.Render()
	return nil
}
