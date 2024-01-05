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
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	addUserToCompanyTemplate = "/api/companies/%s/users"
)

func AddCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Add a user to a company",
		Long:  "Add a user to a company",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = addUserToCompany(cmd.Context(), client, restConfig.CompanyID, options.UserEmail, resources.IAMRole(options.IAMRole))
			cobra.CheckErr(err)
		},
	}

	options.AddNewUserFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("role", resources.IAMRoleCompletion)

	if err != nil {
		// we panic here because if we reach here, something nasty is happening in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func addUserToCompany(ctx context.Context, client *client.APIClient, companyID, userEmail string, role resources.IAMRole) error {
	if !resources.IsValidIAMRole(role) {
		return fmt.Errorf("invalid service account role %s", role)
	}

	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(userEmail) == 0 {
		return fmt.Errorf("the user email is required")
	}

	payload := resources.AddUserRequest{
		Email: userEmail,
		Role:  role,
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Post().
		APIPath(fmt.Sprintf(addUserToCompanyTemplate, companyID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("user %s added to %s company\n", userEmail, companyID)
	return nil
}
