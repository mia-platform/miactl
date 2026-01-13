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

package runtime

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
)

func TestLatestDeployment(t *testing.T) {
	testProjectID := "test-project-id"
	testEnv := "dev"
	now := time.Now().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/deploy/projects/%s/deployment/", testProjectID), r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "1", r.URL.Query().Get("per_page"))
		assert.Equal(t, "success", r.URL.Query().Get("scope"))
		assert.Equal(t, testEnv, r.URL.Query().Get("environment"))

		resp := []resources.DeploymentHistory{
			{
				ID:          "deploy-123",
				Ref:         "main",
				PipelineID:  "pipe-123",
				Status:      "success",
				FinishedAt:  now,
				Environment: testEnv,
			},
		}
		data, _ := resources.EncodeResourceToJSON(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	options := clioptions.NewCLIOptions()
	options.Endpoint = server.URL
	options.ProjectID = testProjectID
	options.Environment = testEnv

	err := runLatestDeployment(t.Context(), options)
	assert.NoError(t, err)
}

func TestLatestDeploymentNoResults(t *testing.T) {
	testProjectID := "test-project-id"
	testEnv := "dev"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	options := clioptions.NewCLIOptions()
	options.Endpoint = server.URL
	options.ProjectID = testProjectID
	options.Environment = testEnv

	err := runLatestDeployment(t.Context(), options)
	assert.NoError(t, err)
}
