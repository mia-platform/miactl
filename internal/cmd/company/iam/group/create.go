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
	"errors"
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	createGroupTemplate = "/api/companies/%s/groups"
)

func AddCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group NAME",
		Short: "Create a new group in a company",
		Long:  "Create a new group in a company",

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			err = createNewGroup(cmd.Context(), client, restConfig.CompanyID, args[0], resources.IAMRole(options.IAMRole))
			cobra.CheckErr(err)
		},
	}

	options.CreateNewGroupFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("role", resources.IAMRoleCompletion(false))

	if err != nil {
		// we panic here because if we reach here, something nasty is happening in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func createNewGroup(ctx context.Context, client *client.APIClient, companyID, groupName string, role resources.IAMRole) error {
	if !resources.IsValidIAMRole(role, false) {
		return fmt.Errorf("invalid service account role %s", role)
	}

	if len(groupName) == 0 {
		return errors.New("a group name is required")
	}

	if len(companyID) == 0 {
		return errors.New("company id is required, please set it via flag or context")
	}

	payload := resources.CreateGroupRequest{
		Name:    groupName,
		Role:    role,
		Members: []string{},
	}

	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Post().
		APIPath(fmt.Sprintf(createGroupTemplate, companyID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	fmt.Printf("group %s added to %s company\n", groupName, companyID)
	return nil
}
