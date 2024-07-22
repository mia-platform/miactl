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

package extensions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/resources/extensibility"
)

const (
	extensibilityAPIPrefix      = "/api/extensibility"
	tenantsExtensionsAPIFmt     = extensibilityAPIPrefix + "/tenants/%s/extensions"
	tenantsExtensionsByIDAPIFmt = tenantsExtensionsAPIFmt + "/%s"
	activationAPIFmt            = tenantsExtensionsByIDAPIFmt + "/activation"
	deactivationAPIFmt          = tenantsExtensionsByIDAPIFmt + "/%s/%s/activation"
)

const IFrameExtensionType = "iframe"

type IE11yClient interface {
	List(ctx context.Context, companyID string) ([]*extensibility.ExtensionInfo, error)
	GetOne(ctx context.Context, companyID string, extensionID string) (*extensibility.ExtensionInfo, error)
	Apply(ctx context.Context, companyID string, extensionData *extensibility.Extension) (string, error)
	Delete(ctx context.Context, companyID string, extensionID string) error
	Activate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error
	Deactivate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error
}

type E11yClient struct {
	c *client.APIClient
}

func New(c *client.APIClient) IE11yClient {
	return &E11yClient{c: c}
}

func (e *E11yClient) List(ctx context.Context, companyID string) ([]*extensibility.ExtensionInfo, error) {
	apiPath := fmt.Sprintf(tenantsExtensionsAPIFmt, companyID)
	resp, err := e.c.Get().APIPath(apiPath).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	if err := e.assertSuccessResponse(resp); err != nil {
		return nil, err
	}

	extensions := make([]*extensibility.ExtensionInfo, 0)
	if err := resp.ParseResponse(&extensions); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	return extensions, nil
}

func (e *E11yClient) GetOne(ctx context.Context, companyID string, extensionID string) (*extensibility.ExtensionInfo, error) {
	apiPath := fmt.Sprintf(tenantsExtensionsByIDAPIFmt, companyID, extensionID)
	resp, err := e.c.Get().APIPath(apiPath).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	if err := e.assertSuccessResponse(resp); err != nil {
		return nil, err
	}

	var extension *extensibility.ExtensionInfo
	if err := resp.ParseResponse(&extension); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}
	return extension, nil
}

type ApplyResponseBody struct {
	ExtensionID string `json:"extensionId"`
}

func (e *E11yClient) Apply(ctx context.Context, companyID string, extensionData *extensibility.Extension) (string, error) {
	apiPath := fmt.Sprintf(tenantsExtensionsAPIFmt, companyID)
	body, err := resources.EncodeResourceToJSON(extensionData)
	if err != nil {
		return "", fmt.Errorf("error serializing request body: %s", err.Error())
	}

	resp, err := e.c.Put().Body(body).APIPath(apiPath).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("error executing request: %w", err)
	}

	if err := e.assertSuccessResponse(resp); err != nil {
		return "", err
	}

	if resp.StatusCode() == http.StatusNoContent {
		return extensionData.ExtensionID, nil
	}

	var applyResult ApplyResponseBody
	if err := resp.ParseResponse(&applyResult); err != nil {
		return "", fmt.Errorf("error parsing response body: %w", err)
	}
	return applyResult.ExtensionID, nil
}

func (e *E11yClient) Delete(ctx context.Context, companyID string, extensionID string) error {
	apiPath := fmt.Sprintf(tenantsExtensionsByIDAPIFmt, companyID, extensionID)
	resp, err := e.c.Delete().APIPath(apiPath).Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	return e.assertSuccessResponse(resp)
}

func (e *E11yClient) Activate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error {
	apiPath := fmt.Sprintf(activationAPIFmt, companyID, extensionID)

	body, err := resources.EncodeResourceToJSON(scope)
	if err != nil {
		return fmt.Errorf("error serializing request body: %s", err.Error())
	}

	resp, err := e.c.
		Post().
		APIPath(apiPath).
		Body(body).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %s", err.Error())
	}
	return e.assertSuccessResponse(resp)
}

func (e *E11yClient) Deactivate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error {
	apiPath := fmt.Sprintf(deactivationAPIFmt, companyID, extensionID, scope.ContextType, scope.ContextID)

	resp, err := e.c.
		Delete().
		APIPath(apiPath).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %s", err.Error())
	}
	return e.assertSuccessResponse(resp)
}

func (e *E11yClient) assertSuccessResponse(resp *client.Response) error {
	if resp.StatusCode() >= http.StatusBadRequest {
		return resp.Error()
	}
	return nil
}
