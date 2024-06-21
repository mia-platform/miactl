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

const extensibilityAPIPrefix = "/api/extensibility"

const (
	listAPIFmt      = extensibilityAPIPrefix + "/tenants/%s/extensions"
	deleteAPIFmt    = extensibilityAPIPrefix + "/tenants/%s/extensions/%s"
	activationAPImt = extensibilityAPIPrefix + "/tenants/%s/extensions/%s/activation"
)

type IE11yClient interface {
	List(ctx context.Context, companyID string) ([]*extensibility.Extension, error)
	Delete(ctx context.Context, companyID string, extensionID string) error
	Activate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error
}

type E11yClient struct {
	c *client.APIClient
}

func New(c *client.APIClient) IE11yClient {
	return &E11yClient{c: c}
}

func (e *E11yClient) List(ctx context.Context, companyID string) ([]*extensibility.Extension, error) {
	apiPath := fmt.Sprintf(listAPIFmt, companyID)
	resp, err := e.c.Get().APIPath(apiPath).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	extensions := make([]*extensibility.Extension, 0)
	if err := resp.ParseResponse(&extensions); err != nil {
		return nil, fmt.Errorf("error parsing response body: %w", err)
	}

	return extensions, nil
}

func (e *E11yClient) Delete(ctx context.Context, companyID string, extensionID string) error {
	apiPath := fmt.Sprintf(deleteAPIFmt, companyID, extensionID)
	resp, err := e.c.Delete().APIPath(apiPath).Do(ctx)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	if resp.StatusCode() >= http.StatusBadRequest {
		return resp.Error()
	}

	return nil
}

func (e *E11yClient) Activate(ctx context.Context, companyID string, extensionID string, scope ActivationScope) error {
	apiPath := fmt.Sprintf(activationAPImt, companyID, extensionID)

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

	if resp.StatusCode() >= http.StatusBadRequest {
		return resp.Error()
	}
	return nil
}
