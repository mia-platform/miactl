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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/spf13/cobra"
)

type basicServiceAccount struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	IssuedAt     int64  `json:"clientIdIssuedAt"`
	Company      string `json:"company"`
}

const (
	companiesURI       = "/api/companies"
	serviceaccountsURI = "/service-accounts"
)

func NewCreateServiceAccountCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create SERVICEACCOUNT [flags]",
		Short: "create a new service account",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := checkCompanyRole(options.CompanyRole)
			if err != nil {
				return err
			}
			currentContext, err := context.GetCurrentContext()
			if err != nil {
				return err
			}
			return context.SetContextValues(cmd, currentContext)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fullURI, err := url.JoinPath(companiesURI, options.CompanyID, serviceaccountsURI)
			if err != nil {
				return err
			}
			mc, err := httphandler.ConfigureDefaultMiaClient(options, fullURI)
			if err != nil {
				return err
			}
			switch options.AuthMethod {
			case "basic":
				credentials, err := createBasicServiceAccount(args[0], mc, options)
				if err != nil {
					return err
				}
				fmt.Printf("Client ID: %s\nClient Secret: %s\n", credentials[0], credentials[1])
			case "jwt":
				// TODO: implement jwt service account creation
				fmt.Println("jwt service accounts are work in progress")
			default:
				return fmt.Errorf("invalid authentication method: it must be one of basic, jwt")
			}
			return nil
		},
	}

	options.AddServiceAccountFlags(cmd)

	return cmd
}

func createBasicServiceAccount(name string, mc *httphandler.MiaClient, opts *clioptions.CLIOptions) ([]string, error) {

	payload := struct {
		Name                    string `json:"name"`
		TokenEndpointAuthMethod string `json:"tokenEndpointAuthMethod"`
		Role                    string `json:"role"`
	}{
		Name:                    name,
		TokenEndpointAuthMethod: "client_secret_basic",
		Role:                    opts.CompanyRole,
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

	var response basicServiceAccount

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return []string{response.ClientID, response.ClientSecret}, nil
}

func checkCompanyRole(role string) error {
	switch role {
	case
		"company-owner",
		"project-admin",
		"maintainer",
		"developer",
		"reporter",
		"guest":
		return nil
	}
	return fmt.Errorf("invalid company role: it must be one of company-owner, project-admin, maintainer, developer, reporter, guest")
}
