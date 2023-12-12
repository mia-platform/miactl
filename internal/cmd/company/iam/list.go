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
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	listAllIAMEntitiesTemplate = "/api/companies/%s/identities"
	GroupsEntityName           = "group"
	UsersEntityName            = "user"
	ServiceAccountsEntityName  = "serviceAccount"
)

func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all IAM entity for a company",
		Long: `A Company can have associated different entities for managing the roles, this command will list
all of them noting the type and the current role associated with them`,

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

	identities := make([]*resources.IAMIdentity, 0)
	if err := resp.ParseResponse(&identities); err != nil {
		return fmt.Errorf("error parsing response body: %w", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Type", "Name", "Roles"})

	caser := cases.Title(language.English)
	for _, identity := range identities {
		table.Append([]string{
			caser.String(readableType(identity.Type)),
			identity.Name,
			caser.String(strings.Join(readableRoles(identity.Roles), ", ")),
		})
	}

	table.Render()
	return nil
}

func readableType(identityType string) string {
	switch identityType {
	case UsersEntityName:
		return "user"
	case GroupsEntityName:
		return "group"
	case ServiceAccountsEntityName:
		return "service account"
	default:
		return identityType
	}
}

func readableRoles(roles []string) []string {
	transformedRoles := make([]string, 0)
	for _, role := range roles {
		switch role {
		case "company-owner":
			transformedRoles = append(transformedRoles, "company owner")
		case "project-admin":
			transformedRoles = append(transformedRoles, "project admin")
		default:
			transformedRoles = append(transformedRoles, role)
		}
	}

	return transformedRoles
}
