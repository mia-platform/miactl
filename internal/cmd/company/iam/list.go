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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all IAM entity for a company",
		Long: `A Company can have associated different entities for managing the roles, this command will list
all of them noting the type and the current role associated with them`,

		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			entityTypes := map[string]bool{
				iam.GroupsEntityName:          options.ShowGroups,
				iam.UsersEntityName:           options.ShowUsers,
				iam.ServiceAccountsEntityName: options.ShowServiceAccounts,
			}

			return listAllIAMEntities(cmd.Context(), client, restConfig.CompanyID, entityTypes, options.Printer())
		},
	}

	options.AddIAMListFlags(cmd.Flags())
	cmd.MarkFlagsMutuallyExclusive("users", "groups", "serviceAccounts")

	cmd.AddCommand(
		listEntity(
			options,
			"users",
			"List all users that have access to the company, directly or via a group",
			"List all users that have access to the company, directly or via a group",
			iam.UsersEntityName,
		),
		listEntity(
			options,
			"groups",
			"List all groups that have access to the company",
			"List all groups that have access to the company",
			iam.GroupsEntityName,
		),
		listEntity(
			options,
			"serviceaccounts",
			"List all service accounts that have access to the company",
			"List all service accounts that have access to the company",
			iam.ServiceAccountsEntityName,
		),
	)

	return cmd
}

func listEntity(options *clioptions.CLIOptions, commandName, shortHelp, longHelp, entityName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName,
		Short: shortHelp,
		Long:  longHelp,

		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			return listSpecificEntities(cmd.Context(), client, restConfig.CompanyID, entityName, options.Printer())
		},
	}

	return cmd
}

func listAllIAMEntities(ctx context.Context, client *client.APIClient, companyID string, entityTypes map[string]bool, p printer.IPrinter) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}
	resp, err := iam.ListAllIAMEntities(ctx, client, companyID, nil, entityTypes)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	rows, err := util.RowsForResources(resp, iam.RowForIAMIdentity)
	if err != nil {
		return err
	}

	p.Keys("ID", "Type", "Name", "Roles").
		BulkRecords(rows...).
		Print()
	return nil
}

func listSpecificEntities(ctx context.Context, client *client.APIClient, companyID string, entityType string, p printer.IPrinter) error {
	if len(companyID) == 0 {
		return fmt.Errorf("missing company id, please set one with the flag or context")
	}
	response, err := iam.ListSpecificEntities(ctx, client, companyID, entityType)

	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := response.Error(); err != nil {
		return err
	}

	var tableHeaders []string
	var rows [][]string
	switch entityType {
	case iam.UsersEntityName:
		tableHeaders = []string{"ID", "Name", "Email", "Roles", "Groups", "Last Login"}
		rows, err = util.RowsForResources(response, iam.RowForUserIdentity)
	case iam.GroupsEntityName:
		tableHeaders = []string{"ID", "Name", "Roles", "Members"}
		rows, err = util.RowsForResources(response, iam.RowForGroupIdentity)
	case iam.ServiceAccountsEntityName:
		tableHeaders = []string{"ID", "Name", "Roles", "Last Login"}
		rows, err = util.RowsForResources(response, iam.RowForServiceAccountIdentity)
	}

	if err != nil {
		return err
	}

	p.Keys(tableHeaders...).BulkRecords(rows...).Print()
	return nil
}
