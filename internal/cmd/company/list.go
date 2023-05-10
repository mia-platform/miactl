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

package company

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/resources"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	companiesURI = "/api/backend/tenants/"
)

// NewListCompaniesCmd func creates a new command
func NewListCompaniesCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list mia companies in the current context",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			currentContext, err := context.GetCurrentContext()
			if err != nil {
				return err
			}
			if err := context.SetContextValues(cmd, currentContext); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mc, err := httphandler.ConfigureDefaultMiaClient(options, companiesURI)
			if err != nil {
				return err
			}
			return listCompanies(mc)
		},
	}
}

// listCompanies retrieves the companies belonging to the current context
func listCompanies(mc *httphandler.MiaClient) error {
	// execute the request
	resp, err := mc.GetSession().Get().ExecuteRequest()
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer resp.Body.Close()

	var companies []resources.Company
	currentContext := mc.GetSession().GetContext()

	if resp.StatusCode == http.StatusOK {
		if err := httphandler.ParseResponseBody(currentContext, resp.Body, &companies); err != nil {
			return fmt.Errorf("error parsing response body: %w", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeader([]string{"Name", "Company ID", "Git Provider", "Pipelines"})
		for _, company := range companies {
			repositoryType := company.Repository.Type
			if repositoryType == "" {
				repositoryType = "gitlab"
			}
			table.Append([]string{company.Name, company.TenantID, repositoryType, company.Pipelines.Type})
		}
		table.Render()
	} else {
		return fmt.Errorf("request failed with status code: %s", resp.Status)
	}

	return nil
}
