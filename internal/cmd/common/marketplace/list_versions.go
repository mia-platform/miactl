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

package marketplace

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

var (
	ErrGenericServerError = errors.New("server error while fetching item versions")
)

func GetItemVersions(ctx context.Context, client *client.APIClient, endpoint string, companyID, itemID string) (*[]marketplace.Release, error) {
	if companyID == "" {
		return nil, marketplace.ErrMissingCompanyID
	}
	resp, err := client.
		Get().
		APIPath(
			fmt.Sprintf(endpoint, companyID, itemID),
		).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		releases := &[]marketplace.Release{}
		err = resp.ParseResponse(releases)
		if err != nil {
			return nil, fmt.Errorf("error parsing response body: %w", err)
		}
		return releases, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s", marketplace.ErrItemNotFound, itemID)
	}
	return nil, ErrGenericServerError
}

func PrintItemVersionList(releases *[]marketplace.Release, p printer.IPrinter) {
	p.Keys("Version", "Name", "Description")

	for _, release := range *releases {
		description := "-"
		if release.Description != "" {
			description = release.Description
		}
		p.Record(
			release.Version,
			release.Name,
			description,
		)
	}
	p.Print()
}
