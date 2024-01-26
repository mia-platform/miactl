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
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	removeMemberTemplate = "/api/companies/%s/groups/%s/members"
)

func RemoveMemberCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-member",
		Short: "Remove one or more users from a group",
		Long:  "Remove one or more users from a company group. The users can be removed via their ids",

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = removeMemberFromGroup(cmd.Context(), client, restConfig.CompanyID, options.EntityID, options.UserIDs)
			cobra.CheckErr(err)
		},
	}

	options.AddRemoveMembersFromGroupFlags(cmd.Flags())
	return cmd
}

func removeMemberFromGroup(ctx context.Context, client *client.APIClient, companyID, groupID string, userIDs []string) error {
	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(groupID) == 0 {
		return fmt.Errorf("a group id is required")
	}

	if len(userIDs) < 1 {
		return fmt.Errorf("at least one user id must be used")
	}

	payload := resources.RemoveMembersToGroup{
		Members: userIDs,
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Delete().
		APIPath(fmt.Sprintf(removeMemberTemplate, companyID, groupID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Println("the users have been removed from the group")
	return nil
}
