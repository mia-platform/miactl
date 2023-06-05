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

package basic

import (
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

type basicServiceAccountResponse struct {
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	ClientIDIssuedAt int64  `json:"clientIdIssuedAt"`
	Company          string `json:"company"`
}

const (
	companyServiceAccountsEndpointTemplate = "/api/companies/%s/service-accounts"
)

func ServiceAccountCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "basic SERVICEACCOUNT [flags]",
		Short: "Create a new basic authentication service account",
		Long: `Create a new basic authentication service account in the provided company or project.

You can create a service account with the same or lower role than the role that
the current authentication has. The role company-owner can be used only when the
service account is created on the company.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceAccountName := args[0]
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			credentials, err := createBasicServiceAccount(client, serviceAccountName, restConfig.CompanyID, resources.ServiceAccountRole(options.ServiceAccountRole))
			if err != nil {
				return err
			}

			cmd.Println("Service account created, please save the following parameters:")
			cmd.Println("")
			cmd.Printf("Client ID: %s\nClient Secret: %s\n", credentials[0], credentials[1])
			return nil
		},
	}

	// add cmd flags
	options.AddServiceAccountFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("service-account-role", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
		// we panic here because if we reach here, something nasty is happenign in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func createBasicServiceAccount(client *client.APIClient, name, companyID string, role resources.ServiceAccountRole) ([]string, error) {
	if !resources.IsValidServiceAccountRole(role) {
		return nil, fmt.Errorf("invalid service account role %s", role)
	}

	payload := &resources.ServiceAccountRequest{
		Name: name,
		Type: resources.ServiceAccountBasic,
		Role: role,
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Post().
		APIPath(fmt.Sprintf(companyServiceAccountsEndpointTemplate, companyID)).
		Body(body).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	response := new(basicServiceAccountResponse)
	if err := resp.ParseResponse(response); err != nil {
		return nil, err
	}

	return []string{response.ClientID, response.ClientSecret}, nil
}
