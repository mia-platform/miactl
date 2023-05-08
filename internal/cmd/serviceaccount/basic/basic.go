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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/spf13/cobra"
)

type basicServiceAccountResponse struct {
	ClientID         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
	ClientIDIssuedAt int64  `json:"clientIdIssuedAt"`
	Company          string `json:"company"`
}

const (
	companyServiceAccountsURITemplate = "api/companies/%s/service-accounts"
)

var validServiceAccountRoles = []string{
	"company-owner",
	"project-admin",
	"maintainer",
	"developer",
	"reporter",
	"guest",
}

func ServiceAccountCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "basic SERVICEACCOUNT [flags]",
		Short: "Create a new basic authentication service account",
		Long: `Create a new basic authentication service account in the provided company or project.

You can create a service account with the same or lower role than the role that
the current authentication has. The role company-owner can be used only when the
service account is created on the company.`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			currentContext, err := context.GetCurrentContext()
			if err != nil {
				return err
			}
			return context.SetContextValues(cmd, currentContext)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fullURI := fmt.Sprintf(companyServiceAccountsURITemplate, options.CompanyID)
			mc, err := httphandler.ConfigureDefaultMiaClient(options, fullURI)
			if err != nil {
				return err
			}
			credentials, err := createBasicServiceAccount(args[0], mc, options)
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
		return validServiceAccountRoles, cobra.ShellCompDirectiveDefault
	})

	if err != nil {
		// we panic here because if we reach here, something nasty is happenign in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func createBasicServiceAccount(name string, mc *httphandler.MiaClient, opts *clioptions.CLIOptions) ([]string, error) {
	if !isValidServiceAccountRole(opts.ServiceAccountRole) {
		return []string{}, fmt.Errorf("invalid service account role %s", opts.ServiceAccountRole)
	}

	payload := struct {
		Name                    string `json:"name"`
		TokenEndpointAuthMethod string `json:"tokenEndpointAuthMethod"`
		Role                    string `json:"role"`
	}{
		Name:                    name,
		TokenEndpointAuthMethod: "client_secret_basic",
		Role:                    opts.ServiceAccountRole,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error building request payload: %w", err)
	}

	resp, err := mc.SessionHandler.Post(bytes.NewBuffer(jsonPayload)).ExecuteRequest()
	if err != nil {
		return nil, fmt.Errorf("error executing request for service account creation: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// TODO: review console error when creating a service account with a name that is already taken
		// map[message:POST http://client-credentials/clients: 400 - {"error":"invalid_client_metadata",
		// "error_description":"fails to create the client"} statusCode:400]
		return nil, fmt.Errorf("service account creation failed with status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var response basicServiceAccountResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return []string{response.ClientID, response.ClientSecret}, nil
}

func isValidServiceAccountRole(role string) bool {
	for _, validRole := range validServiceAccountRoles {
		if validRole == role {
			return true
		}
	}

	return false
}
