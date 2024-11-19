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

package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"
)

const (
	tenantsAPIPrefix     = "/api/backend/tenants/"
	projectsAPIPrefix    = "/api/backend/projects/"
	getProjectAPIFmt     = projectsAPIPrefix + "%s"
	patchTenantRulesFmt  = tenantsAPIPrefix + "%s/rules"
	patchProjectRulesFmt = projectsAPIPrefix + "%s/rules"
)

type IRulesClient interface {
	ListTenantRules(ctx context.Context, companyID string) ([]*rulesentities.SaveChangesRules, error)
	ListProjectRules(ctx context.Context, projectID string) ([]*rulesentities.ProjectSaveChangesRules, error)
	UpdateTenantRules(ctx context.Context, companyID string, rules []*rulesentities.SaveChangesRules) error
	UpdateProjectRules(ctx context.Context, projectId string, rules []*rulesentities.SaveChangesRules) error
}

type RulesClient struct {
	c *client.APIClient
}

func New(c *client.APIClient) IRulesClient {
	return &RulesClient{c: c}
}

func (e *RulesClient) ListTenantRules(ctx context.Context, companyID string) ([]*rulesentities.SaveChangesRules, error) {
	request := e.c.Get().APIPath(tenantsAPIPrefix)
	request.SetParam("search", companyID)

	resp, err := request.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	if err := e.assertSuccessResponse(resp); err != nil {
		return nil, err
	}

	var tenants []resources.Company
	if err := resp.ParseResponse(&tenants); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}
	if len(tenants) == 0 {
		return nil, fmt.Errorf("Company %s not found", companyID)
	}
	var tenant *resources.Company
	for _, possible := range tenants {
		if possible.TenantID == companyID {
			tenant = &possible
			break
		}
	}
	if tenant == nil {
		return nil, fmt.Errorf("Company %s not found", companyID)
	}
	if len(tenant.ConfigurationManagement.SaveChangesRules) == 0 {
		return []*rulesentities.SaveChangesRules{}, nil
	}

	return tenant.ConfigurationManagement.SaveChangesRules, nil
}

func (e *RulesClient) ListProjectRules(ctx context.Context, projectID string) ([]*rulesentities.ProjectSaveChangesRules, error) {
	request := e.c.Get().APIPath(fmt.Sprintf(getProjectAPIFmt, projectID))
	request.SetParam("withTenant", "true")

	resp, err := request.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	if err := e.assertSuccessResponse(resp); err != nil {
		return nil, err
	}

	var project resources.Project
	if err := resp.ParseResponse(&project); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	if len(project.ConfigurationManagement.SaveChangesRules) == 0 {
		return []*rulesentities.ProjectSaveChangesRules{}, nil
	}

	return project.ConfigurationManagement.SaveChangesRules, nil
}

type UpdateRequestBody struct {
	ConfigurationManagement *resources.ConfigurationManagement `json:"configurationManagement"`
}

func (e *RulesClient) UpdateTenantRules(ctx context.Context, companyID string, rules []*rulesentities.SaveChangesRules) error {
	requestBody := UpdateRequestBody{
		ConfigurationManagement: &resources.ConfigurationManagement{
			SaveChangesRules: rules,
		},
	}
	bodyData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request := e.c.Patch().
		APIPath(
			fmt.Sprintf(patchTenantRulesFmt, companyID),
		).
		Body(bodyData)

	resp, err := request.Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	if err := e.assertSuccessResponse(resp); err != nil {
		return err
	}

	return nil
}

func (e *RulesClient) UpdateProjectRules(ctx context.Context, projectID string, rules []*rulesentities.SaveChangesRules) error {
	requestBody := UpdateRequestBody{
		ConfigurationManagement: &resources.ConfigurationManagement{
			SaveChangesRules: rules,
		},
	}
	bodyData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request := e.c.Patch().
		APIPath(
			fmt.Sprintf(patchProjectRulesFmt, projectID),
		).
		Body(bodyData)

	resp, err := request.Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	if err := e.assertSuccessResponse(resp); err != nil {
		return err
	}

	return nil
}

func (e *RulesClient) assertSuccessResponse(resp *client.Response) error {
	if resp.StatusCode() >= http.StatusBadRequest {
		return resp.Error()
	}
	return nil
}
