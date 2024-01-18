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

package user

import (
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

func EditCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Edit a user in a company",
		Long:  "Edit a user in a company",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = editCompanyUser(cmd.Context(), client, restConfig.CompanyID, options.UserID, resources.IAMRole(options.IAMRole))
			cobra.CheckErr(err)
		},
	}

	options.AddEditUserFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("role", resources.IAMRoleCompletion)

	if err != nil {
		// we panic here because if we reach here, something nasty is happening in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func editCompanyUser(ctx context.Context, client *client.APIClient, companyID, userID string, role resources.IAMRole) error {
	if !resources.IsValidIAMRole(role) {
		return fmt.Errorf("invalid service account role %s", role)
	}

	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(userID) == 0 {
		return fmt.Errorf("the user id is required")
	}

	payload := resources.EditIAMRole{
		Role: role,
	}

	resp, err := iam.EditUserRole(ctx, client, companyID, userID, payload)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("user %s role successfully updated\n", userID)
	return nil
}
