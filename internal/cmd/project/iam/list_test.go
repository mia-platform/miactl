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

package iam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/iam"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllIAMIdentities(t *testing.T) {
	companyID := "company"
	projectID := "project"
	testCases := map[string]struct {
		server       *httptest.Server
		searchParams map[string]bool
		err          bool
	}{
		"valid get response": {
			server:       iam.TestServerForProjectIAMList(t, companyID, projectID),
			searchParams: map[string]bool{},
		},
		"valid get with search parameters": {
			server: iam.TestServerForProjectIAMList(t, companyID, projectID),
			searchParams: map[string]bool{
				iam.ServiceAccountsEntityName: true,
				iam.GroupsEntityName:          true,
			},
		},
		"invalid body response": {
			server: iam.ErrorTestServerForProjectIAMList(t, companyID, projectID),
			err:    true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			clientConfig := &client.Config{
				Transport: http.DefaultTransport,
				Host:      testCase.server.URL,
			}
			client, err := client.APIClientForConfig(clientConfig)
			require.NoError(t, err)
			err = listAllIAMEntities(context.TODO(), client, companyID, projectID, testCase.searchParams, &printer.NopPrinter{})
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
