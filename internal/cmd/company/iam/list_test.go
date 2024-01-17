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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllIAMIdentities(t *testing.T) {
	companyID := "company-id"
	testCases := map[string]struct {
		server       *httptest.Server
		companyID    string
		searchParams map[string]bool
		err          bool
	}{
		"valid get response": {
			server:       iam.TestServerForCompanyIAMList(t, companyID),
			companyID:    companyID,
			searchParams: map[string]bool{},
		},
		"valid get with search parameters": {
			server:    iam.TestServerForCompanyIAMList(t, companyID),
			companyID: companyID,
			searchParams: map[string]bool{
				iam.ServiceAccountsEntityName: true,
				iam.GroupsEntityName:          true,
			},
		},
		"invalid body response": {
			server:    iam.ErrorTestServerForCompanyIAM(t, companyID),
			companyID: companyID,
			err:       true,
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
			err = listAllIAMEntities(context.TODO(), client, testCase.companyID, testCase.searchParams)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListUsersIdentities(t *testing.T) {
	companyID := "user"
	testCases := map[string]struct {
		server    *httptest.Server
		companyID string
		err       bool
	}{
		"valid get response": {
			server:    iam.TestServerForCompanySpecificList(t, companyID, iam.UsersEntityName),
			companyID: companyID,
		},
		"invalid body response": {
			server:    iam.ErrorTestServerForCompanyIAM(t, companyID),
			companyID: companyID,
			err:       true,
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

			err = listSpecificEntities(context.TODO(), client, testCase.companyID, iam.UsersEntityName)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListGroupsIdentities(t *testing.T) {
	companyID := "group-list"
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companyID    string
		err          bool
	}{
		"valid get response": {
			server:    iam.TestServerForCompanySpecificList(t, companyID, iam.GroupsEntityName),
			companyID: companyID,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
		},
		"invalid body response": {
			server:    iam.ErrorTestServerForCompanyIAM(t, companyID),
			companyID: companyID,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)

			err = listSpecificEntities(context.TODO(), client, testCase.companyID, iam.GroupsEntityName)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceAccountGroupsIdentities(t *testing.T) {
	companyID := "service-account"
	testCases := map[string]struct {
		server       *httptest.Server
		clientConfig *client.Config
		companyID    string
		err          bool
	}{
		"valid get response": {
			server:    iam.TestServerForCompanySpecificList(t, companyID, iam.ServiceAccountsEntityName),
			companyID: companyID,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
		},
		"invalid body response": {
			server:    iam.ErrorTestServerForCompanyIAM(t, companyID),
			companyID: companyID,
			clientConfig: &client.Config{
				Transport: http.DefaultTransport,
			},
			err: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			defer testCase.server.Close()
			testCase.clientConfig.Host = testCase.server.URL
			client, err := client.APIClientForConfig(testCase.clientConfig)
			require.NoError(t, err)

			err = listSpecificEntities(context.TODO(), client, testCase.companyID, iam.ServiceAccountsEntityName)
			if testCase.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
