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

package resources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
)

var successJob = []resources.Job{
	{
		Name:           "new-job-name-wait-success",
		Succeeded:      1,
		Active:         0,
		Failed:         0,
		Age:            time.Now().Add(-5 * time.Second),
		StartTime:      time.Now().Add(-5 * time.Second),
		CompletionTime: time.Now().Add(-4 * time.Second),
	},
}

var failedJob = []resources.Job{
	{
		Name:           "new-job-name-wait-fail",
		Succeeded:      0,
		Active:         0,
		Failed:         1,
		Age:            time.Now().Add(-5 * time.Second),
		StartTime:      time.Now().Add(-5 * time.Second),
		CompletionTime: time.Now().Add(-4 * time.Second),
	},
}

var successPod = []resources.Pod{
	{
		Name:   "new-job-name-wait-success-pod",
		Phase:  "Succeeded",
		Status: "ok",
		Age:    time.Now().Add(-5 * time.Second),
		Component: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			{
				Name:    "new-job-name-wait-success-pod",
				Version: "1.0.0",
			},
		},
		Containers: []struct {
			Name         string `json:"name"`
			Ready        bool   `json:"ready"`
			RestartCount int    `json:"restartCount"`
			Status       string `json:"status"`
		}{
			{
				Name:         "new-job-name-wait-success-pod-container",
				Ready:        false,
				RestartCount: 0,
				Status:       "Completed",
			},
		},
		Labels: map[string]string{
			"job-name": "new-job-name-wait-success",
		},
	},
}

var failedPod = []resources.Pod{
	{
		Name:   "new-job-name-wait-success-pod",
		Phase:  "Failed",
		Status: "ko",
		Age:    time.Now().Add(-5 * time.Second),
		Component: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			{
				Name:    "new-job-name-wait-success-pod",
				Version: "1.0.0",
			},
		},
		Containers: []struct {
			Name         string `json:"name"`
			Ready        bool   `json:"ready"`
			RestartCount int    `json:"restartCount"`
			Status       string `json:"status"`
		}{
			{
				Name:         "new-job-name-wait-success-pod-container",
				Ready:        false,
				RestartCount: 0,
				Status:       "Completed",
			},
		},
		Labels: map[string]string{
			"job-name": "new-job-name-wait-success",
		},
	},
}

func TestCreateJob(t *testing.T) {
	testCases := map[string]struct {
		testServer            *httptest.Server
		projectID             string
		environment           string
		waitJobCompletion     bool
		waitJobTimeoutSeconds int
		err                   bool
	}{
		"create job end with success": {
			testServer:            createJobTestServer(t),
			projectID:             "success",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 600,
		},
		"create job end with error": {
			testServer:            createJobTestServer(t),
			projectID:             "fail",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 600,
			err:                   true,
		},
		"fail if no project id": {
			testServer:            createJobTestServer(t),
			projectID:             "",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 600,
			err:                   true,
		},
		"fail if environment": {
			testServer:            createJobTestServer(t),
			projectID:             "success",
			environment:           "",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 600,
			err:                   true,
		},
		"create job and wait for completion with success": {
			testServer:            createJobTestServer(t),
			projectID:             "success",
			environment:           "env-id-wait",
			waitJobCompletion:     true,
			waitJobTimeoutSeconds: 600,
		},
		"create job and wait for completion with failure - timeout": {
			testServer:            createJobTestServer(t),
			projectID:             "timeout",
			environment:           "env-id-wait",
			waitJobCompletion:     true,
			waitJobTimeoutSeconds: 15,
			err:                   true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testCase.testServer
			defer server.Close()

			client, err := client.APIClientForConfig(&client.Config{
				Host: server.URL,
			})
			require.NoError(t, err)

			err = createJob(t.Context(), client, testCase.projectID, testCase.environment, "cronjob-name", testCase.waitJobCompletion, testCase.waitJobTimeoutSeconds)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createJobTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "success", "env-id"):
			response := resources.CreateJob{
				JobName: "new-job-name",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "fail", "env-id"):
			response := resources.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Error creating job",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "success", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-wait-success",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "fail", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-wait-fail",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "timeout", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-wait-timeout",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "success", "env-id-wait"):
			response := successJob
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "fail", "env-id-wait"):
			response := failedJob
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "timeout", "env-id-wait"):
			response := []resources.Job{
				{
					Name:      "new-job-name-wait-timeout",
					Succeeded: 0,
					Active:    1,
					Failed:    0,
					Age:       time.Now().Add(-5 * time.Second),
					StartTime: time.Now().Add(-5 * time.Second),
				},
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describePodsTemplate, "success", "env-id-wait"):
			response := successPod
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describePodsTemplate, "fail", "env-id-wait"):
			response := failedPod
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describePodsTemplate, "timeout", "env-id-wait"):
			response := []resources.Pod{
				{
					Name:   "new-job-name-wait-timeout-pod",
					Phase:  "Running",
					Status: "Running",
					Age:    time.Now().Add(-5 * time.Second),
					Labels: map[string]string{
						"job-name": "new-job-name-wait-timeout",
					},
				},
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}
