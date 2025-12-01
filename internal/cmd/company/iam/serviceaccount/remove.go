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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
)

const (
	removeServiceAccountTemplate = "/api/companies/%s/service-accounts/%s"
)

func RemoveCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceaccount",
		Short: "Remove a service account from a company",
		Long:  "Remove a service account from a company",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = removeCompanyServiceAccount(cmd.Context(), client, restConfig.CompanyID, options.ServiceAccountID)
			cobra.CheckErr(err)
		},
	}

	options.AddRemoveServiceAccountFlags(cmd.Flags())
	return cmd
}

func removeCompanyServiceAccount(ctx context.Context, client *client.APIClient, companyID, serviceAccountID string) error {
	if len(companyID) == 0 {
		return errors.New("company id is required, please set it via flag or context")
	}

	if len(serviceAccountID) == 0 {
		return errors.New("the service account id is required")
	}

	request := client.
		Delete().
		APIPath(fmt.Sprintf(removeServiceAccountTemplate, companyID, serviceAccountID))

	resp, err := request.Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("service account %s successfully removed\n", serviceAccountID)
	return nil
}
