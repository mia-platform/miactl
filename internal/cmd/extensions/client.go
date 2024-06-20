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

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources/extensibility"
)

const (
	listAPIFmt = "/api/extensibility/tenants/%s/extensions"
)

type IE11yClient interface {
	List(ctx context.Context, companyID string) ([]*extensibility.Extension, error)
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
