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
	addMemberTemplate = "/api/companies/%s/groups/%s/members"
)

func AddMemberCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-member",
		Short: "Add one or more users to a group",
		Long:  "Add one or more users to a company group. The users can be added via their emails",

		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			return addMemberToGroup(cmd.Context(), client, restConfig.CompanyID, options.GroupID, options.UserEmails)
		},
	}

	options.AddMemberToGroupFlags(cmd.Flags())
	return cmd
}

func addMemberToGroup(ctx context.Context, client *client.APIClient, companyID, groupID string, userEmails []string) error {
	if len(companyID) == 0 {
		return fmt.Errorf("company id is required, please set it via flag or context")
	}

	if len(groupID) == 0 {
		return fmt.Errorf("a group id is required")
	}

	if len(userEmails) < 1 {
		return fmt.Errorf("at least one user must be added to the group")
	}

	payload := resources.AddMembersToGroup{
		Members: userEmails,
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Post().
		APIPath(fmt.Sprintf(addMemberTemplate, companyID, groupID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Println("the users has been added to the group")
	return nil
}
