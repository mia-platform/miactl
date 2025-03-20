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

package project

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	successConversionBody = `{
	"services": {
		"service1": {}
	},
	"configMaps": {
		"configmap1": {}
	},
	"secrets": {
		"secret1": {}
	},
	"errors": [
		"error line1",
		"error line2"
	]
}`
	successConversionBodyWithoutError = `{
	"services": {
		"service1": {}
	},
	"configMaps": {
		"configmap1": {}
	},
	"secrets": {
		"secret1": {}
	},
	"errors": []
}`
	successConversionBodyWithServiceACcounts = `{
	"services": {
		"service1": {}
	},
	"configMaps": {
		"configmap1": {}
	},
	"secrets": {
		"secret1": {}
	},
	"serviceAccounts": {
		"serviceAccount1": {}
	},
	"errors": []
	}`
	emptyConversionBody = `{
	"services": {},
	"configMaps": {},
	"secrets": {},
	"errors": []
}`
	getConfigurationBody = `{
	"endpoints": {},
	"collections": {},
	"groups": [],
	"secrets": [],
	"cmsCategories": {},
	"cmsSettings": {
			"accessGroupsExpression": ""
	},
	"cmsAnalytics": {},
	"cmsDashboard": [],
	"decorators": {
			"preDecorators": {},
			"postDecorators": {}
	},
	"services": {},
	"applications": {},
	"listeners": {},
	"apiVersions": [],
	"version": "version",
	"platformVersion": "platform-version",
	"lastConfigFileCommitId": "00000000-0000-0000-0000-000000000000",
	"lastCommitAuthor": "Jhon Smith",
	"commitId": "00000000-0000-0000-0000-000000000000",
	"committedDate": "1970-01-01T00:00:00.000Z",
	"configMaps": {},
	"serviceSecrets": {},
	"unsecretedVariables": [],
	"fastDataConfig": {},
	"enabledFeatures": {}
}`
	notEmptyConfigurationBody = `{
	"endpoints": {},
	"collections": {},
	"groups": [],
	"secrets": [],
	"cmsCategories": {},
	"cmsSettings": {
			"accessGroupsExpression": ""
	},
	"cmsAnalytics": {},
	"cmsDashboard": [],
	"decorators": {
			"preDecorators": {},
			"postDecorators": {}
	},
	"services": {
		"a-service-name": {}
	},
	"applications": {},
	"listeners": {},
	"apiVersions": [],
	"version": "version",
	"platformVersion": "platform-version",
	"lastConfigFileCommitId": "00000000-0000-0000-0000-000000000000",
	"lastCommitAuthor": "Jhon Smith",
	"commitId": "00000000-0000-0000-0000-000000000000",
	"committedDate": "1970-01-01T00:00:00.000Z",
	"configMaps": {
		"configmapname": {}
	},
	"serviceSecrets": {},
	"unsecretedVariables": [],
	"fastDataConfig": {},
	"extensionsConfig": {
			"files": {}
	},
	"enabledFeatures": {}
}`
	modifiedConfigurationBody = `{"config":{"apiVersions":[],"applications":{},"cmsAnalytics":{},"cmsCategories":{},"cmsDashboard":[],"cmsSettings":{"accessGroupsExpression":""},"collections":{},"commitId":"00000000-0000-0000-0000-000000000000","configMaps":{"configmap1":{}},"decorators":{"postDecorators":{},"preDecorators":{}},"enabledFeatures":{},"endpoints":{},"groups":[],"lastConfigFileCommitId":"00000000-0000-0000-0000-000000000000","listeners":{},"secrets":[],"serviceSecrets":{"secret1":{}},"services":{"service1":{}},"unsecretedVariables":[],"version":"version"},"deletedElements":{},"extensionsConfig":{"files":{}},"fastDataConfig":{},"microfrontendPluginsConfig":{},"previousSave":"00000000-0000-0000-0000-000000000000","title":"[CLI] Import resource from kubernetes"}
`
	modifiedConfigurationBodyWithServiceAccount = `{"config":{"apiVersions":[],"applications":{},"cmsAnalytics":{},"cmsCategories":{},"cmsDashboard":[],"cmsSettings":{"accessGroupsExpression":""},"collections":{},"commitId":"00000000-0000-0000-0000-000000000000","configMaps":{"configmap1":{}},"decorators":{"postDecorators":{},"preDecorators":{}},"enabledFeatures":{},"endpoints":{},"groups":[],"lastConfigFileCommitId":"00000000-0000-0000-0000-000000000000","listeners":{},"secrets":[],"serviceAccounts":{"serviceAccount1":{}},"serviceSecrets":{"secret1":{}},"services":{"service1":{}},"unsecretedVariables":[],"version":"version"},"deletedElements":{},"extensionsConfig":{"files":{}},"fastDataConfig":{},"microfrontendPluginsConfig":{},"previousSave":"00000000-0000-0000-0000-000000000000","title":"[CLI] Import resource from kubernetes"}
`
)

func TestCmdCreation(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, ImportCmd(clioptions.NewCLIOptions()))
}

func TestImportValidation(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	assert.ErrorContains(t, importResources(ctx, nil, "", "", "", nil), "missing project id")
	assert.ErrorContains(t, importResources(ctx, nil, "projectID", "", "", nil), "missing revision")
	assert.ErrorContains(t, importResources(ctx, nil, "projectID", "revision", "", nil), "missing file path")
}

func TestImportResources(t *testing.T) {
	t.Parallel()
	testdata := "testdata"
	projectID := "projectID"
	revision := "revision"

	tests := map[string]struct {
		inputPath     string
		testServer    *httptest.Server
		outputText    string
		expectedError string
	}{
		"successfully import services": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					_, err := w.Write([]byte(getConfigurationBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodPost:
					defer r.Body.Close()
					bodyData, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, modifiedConfigurationBody, string(bodyData))
					_, err = w.Write([]byte(`{}`))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			outputText: "Configuration imported successfully with warnings:\n	- error line1\n	- error line2\n",
		},
		"successfully import services without errors": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBodyWithoutError))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					_, err := w.Write([]byte(getConfigurationBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodPost:
					defer r.Body.Close()
					bodyData, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, modifiedConfigurationBody, string(bodyData))
					_, err = w.Write([]byte(`{}`))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			outputText: "Configuration imported successfully\n",
		},
		"successfully import services with service accounts": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBodyWithServiceACcounts))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					_, err := w.Write([]byte(getConfigurationBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodPost:
					defer r.Body.Close()
					bodyData, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, modifiedConfigurationBodyWithServiceAccount, string(bodyData))
					_, err = w.Write([]byte(`{}`))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			outputText: "Configuration imported successfully\n",
		},
		"empty converted resource return early": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(emptyConversionBody))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			outputText: "No valid resources found to import\n",
		},
		"failure to read resources": {
			inputPath: filepath.Join(testdata, "missing-folder"),
			testServer: importTestServer(t, func(_ http.ResponseWriter, _ *http.Request) bool {
				return false
			}),
			expectedError: "no such file or directory",
		},
		"failure to convert resources": {
			inputPath: filepath.Join(testdata, "multiple-files"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost {
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte(`{"message": "bad request", "statusCode": 400}`))
					require.NoError(t, err)
					return true
				}
				return false
			}),
			expectedError: "bad request",
		},
		"failure to parse imported resources": {
			inputPath: filepath.Join(testdata, "multiple-files"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost {
					_, err := w.Write([]byte(`{"message": "bad request"`))
					require.NoError(t, err)
					return true
				}
				return false
			}),
			expectedError: "cannot parse server response",
		},
		"failure to get configuration": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					w.WriteHeader(http.StatusNotFound)
					_, err := w.Write([]byte(`{"message": "not found", "statusCode": 404}`))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			expectedError: "not found",
		},
		"failure to save configuration": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					_, err := w.Write([]byte(getConfigurationBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodPost:
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte(`{"statusCode": 400, "message": "bad request"}`))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			expectedError: "bad request",
		},
		"failure if configuration is not empty": {
			inputPath: filepath.Join(testdata, "single-file.yaml"),
			testServer: importTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
				switch {
				case r.URL.Path == fmt.Sprintf(convertEndpointTemplate, projectID) && r.Method == http.MethodPost:
					_, err := w.Write([]byte(successConversionBody))
					require.NoError(t, err)
				case r.URL.Path == fmt.Sprintf(configurationEndpointTemplate, projectID, revision) && r.Method == http.MethodGet:
					_, err := w.Write([]byte(notEmptyConfigurationBody))
					require.NoError(t, err)
				default:
					return false
				}

				return true
			}),
			expectedError: "cannot import services in a non empty project",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			server := testCase.testServer
			defer server.Close()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			outputBuffer := bytes.NewBuffer([]byte{})

			err = importResources(ctx, client, projectID, revision, testCase.inputPath, outputBuffer)

			if len(testCase.expectedError) > 0 {
				require.ErrorContains(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.outputText, outputBuffer.String())
		})
	}
}

func importTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler(w, r) {
			return
		}

		t.Logf("unexpected request: %#v\n%#v", r.URL, r)
		w.WriteHeader(http.StatusNotFound)
		assert.Fail(t, "unexpected request")
	}))
}
