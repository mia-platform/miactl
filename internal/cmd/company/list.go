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
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	listCompaniesEndpoint = "/api/backend/tenants/"
)

// ListCmd return a new cobra command for listing companies
func ListCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List user companies",
		Long: `List the companies that the current user can access.

Companies can be used to logically group projects by organizations or internal teams.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			printer := options.Printer()
			return listCompanies(cmd.Context(), client, printer)
		},
	}
}

// listCompanies retrieves the companies belonging to the current context
func listCompanies(ctx context.Context, client *client.APIClient, p printer.IPrinter) error {
	// execute the request
	resp, err := client.Get().APIPath(listCompaniesEndpoint).Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if err := resp.Error(); err != nil {
		return err
	}

	companies := make([]*resources.Company, 0)
	if err := resp.ParseResponse(&companies); err != nil {
		return fmt.Errorf("error parsing response body: %w", err)
	}

	p.Keys("Name", "Company ID", "Git Provider", "Pipelines")
	for _, company := range companies {
		repositoryType := company.Repository.Type
		if repositoryType == "" {
			repositoryType = "gitlab"
		}
		p.Record(company.Name, company.TenantID, repositoryType, company.Pipelines.Type)
	}

	p.Print()
	return nil
}
