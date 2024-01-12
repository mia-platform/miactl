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

package group

import (
	"context"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

const (
	removeGroupTemplate = "/api/companies/%s/groups/%s"
)

func RemoveCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Remove a group from a company",
		Long:  "Remove a group from a company",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = removeCompanyGroup(cmd.Context(), client, restConfig.CompanyID, options.GroupID)
			cobra.CheckErr(err)
		},
	}

	options.AddRemoveGroupFlags(cmd.Flags())
	return cmd
}

func removeCompanyGroup(ctx context.Context, client *client.APIClient, companyID, groupID string) error {
	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(groupID) == 0 {
		return fmt.Errorf("the group id is required")
	}

	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(removeGroupTemplate, companyID, groupID)).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("group %s successfully removed\n", groupID)
	return nil
}
