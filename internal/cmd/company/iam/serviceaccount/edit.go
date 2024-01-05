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

package serviceaccount

import (
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	editServiceAccountRoleTemplate = "/api/companies/%s/service-accounts/%s"
)

func EditCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceaccount",
		Short: "Edit a service account in a company",
		Long:  "Edit a service account in a company",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = editCompanyServiceAccount(cmd.Context(), client, restConfig.CompanyID, options.ServiceAccountID, resources.ServiceAccountRole(options.IAMRole))
			cobra.CheckErr(err)
		},
	}

	options.AddEditServiceAccountFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("role", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			resources.ServiceAccountRoleGuest.String(),
			resources.ServiceAccountRoleReporter.String(),
			resources.ServiceAccountRoleDeveloper.String(),
			resources.ServiceAccountRoleMaintainer.String(),
			resources.ServiceAccountRoleProjectAdmin.String(),
			resources.ServiceAccountRoleCompanyOwner.String(),
		}, cobra.ShellCompDirectiveDefault
	})

	if err != nil {
		// we panic here because if we reach here, something nasty is happening in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func editCompanyServiceAccount(ctx context.Context, client *client.APIClient, companyID, serviceAccountID string, role resources.ServiceAccountRole) error {
	if !resources.IsValidServiceAccountRole(role) {
		return fmt.Errorf("invalid service account role %s", role)
	}

	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(serviceAccountID) == 0 {
		return fmt.Errorf("the user id is required")
	}

	payload := resources.AddUserRequest{
		Role: role,
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Patch().
		APIPath(fmt.Sprintf(editServiceAccountRoleTemplate, companyID, serviceAccountID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("service account %s role successfully updated\n", serviceAccountID)
	return nil
}
