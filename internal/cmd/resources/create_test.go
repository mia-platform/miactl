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
	"context"
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
			testServer:            createJobTestServer(t, nil),
			projectID:             "success",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 1,
		},
		"create job end with error": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "fail",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 1,
			err:                   true,
		},
		"fail if no project id": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "",
			environment:           "env-id",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 1,
			err:                   true,
		},
		"fail if environment": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "success",
			environment:           "",
			waitJobCompletion:     false,
			waitJobTimeoutSeconds: 1,
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

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			err = createJob(ctx, client, testCase.projectID, testCase.environment, "cronjob-name", testCase.waitJobCompletion, testCase.waitJobTimeoutSeconds)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWaitForJobCompletion(t *testing.T) {
	testCases := map[string]struct {
		testServer            *httptest.Server
		projectID             string
		environment           string
		jobName               string
		waitJobTimeoutSeconds int
		err                   bool
	}{
		"wait for completion with success": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "success",
			environment:           "env-id-wait",
			jobName:               "new-job-name-wait-success",
			waitJobTimeoutSeconds: 5,
		},
		"wait for completion with timeout": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "timeout",
			environment:           "env-id-wait",
			jobName:               "new-job-name-wait-timeout",
			waitJobTimeoutSeconds: 1,
			err:                   true,
		},
		"wait - job not found": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "not-found",
			environment:           "env-id-wait",
			jobName:               "new-job-name-not-found",
			waitJobTimeoutSeconds: 5,
			err:                   true,
		},
		"wait - retry on error": {
			testServer: createJobTestServer(t, &serverState{
				jobStatusCallCount: 0,
			}),
			projectID:             "retry",
			environment:           "env-id-wait",
			jobName:               "new-job-name-retry",
			waitJobTimeoutSeconds: 5,
		},
		"wait - max retries exceeded": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "max-retry",
			environment:           "env-id-wait",
			jobName:               "new-job-name-max-retry",
			waitJobTimeoutSeconds: 5,
			err:                   true,
		},
		"wait - pods retrieval error": {
			testServer:            createJobTestServer(t, nil),
			projectID:             "pods-error",
			environment:           "env-id-wait",
			jobName:               "new-job-name-pods-error",
			waitJobTimeoutSeconds: 5,
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

			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			err = waitForJobCompletionWithInterval(ctx, client, testCase.projectID, testCase.environment, testCase.jobName, testCase.waitJobTimeoutSeconds, 100*time.Millisecond)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type serverState struct {
	jobStatusCallCount int
}

func createJobTestServer(t *testing.T, state *serverState) *httptest.Server {
	t.Helper()
	if state == nil {
		state = &serverState{}
	}
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
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "not-found", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-not-found",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "retry", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-retry",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "max-retry", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-max-retry",
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf(createJobTemplate, "pods-error", "env-id-wait"):
			response := resources.CreateJob{
				JobName: "new-job-name-pods-error",
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
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "not-found", "env-id-wait"):
			response := []resources.Job{}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "retry", "env-id-wait"):
			state.jobStatusCallCount++
			if state.jobStatusCallCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			response := []resources.Job{
				{
					Name:      "new-job-name-retry",
					Succeeded: 1,
					Active:    0,
					Failed:    0,
					Age:       time.Now().Add(-5 * time.Second),
					StartTime: time.Now().Add(-5 * time.Second),
				},
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "max-retry", "env-id-wait"):
			w.WriteHeader(http.StatusInternalServerError)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describeJobsTemplate, "pods-error", "env-id-wait"):
			response := []resources.Job{
				{
					Name:      "new-job-name-pods-error",
					Succeeded: 1,
					Active:    0,
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
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describePodsTemplate, "retry", "env-id-wait"):
			response := []resources.Pod{
				{
					Name:   "new-job-name-retry-pod",
					Phase:  "Succeeded",
					Status: "ok",
					Age:    time.Now().Add(-5 * time.Second),
					Labels: map[string]string{
						"job-name": "new-job-name-retry",
					},
				},
			}
			data, err := resources.EncodeResourceToJSON(response)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf(describePodsTemplate, "pods-error", "env-id-wait"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
			assert.Failf(t, "unexpected http call", "received call with method: %s uri %s", r.Method, r.RequestURI)
		}
	}))
	return server
}

func TestGetJobPodsFromDescribe(t *testing.T) {
	testCases := map[string]struct {
		pods        []resources.Pod
		jobName     string
		expectedLen int
	}{
		"returns pods for matching job": {
			pods: []resources.Pod{
				{
					Name: "pod-1",
					Labels: map[string]string{
						"job-name": "test-job",
					},
					Age: time.Now().Add(-10 * time.Second),
				},
				{
					Name: "pod-2",
					Labels: map[string]string{
						"job-name": "test-job",
					},
					Age: time.Now().Add(-5 * time.Second),
				},
				{
					Name: "pod-3",
					Labels: map[string]string{
						"job-name": "other-job",
					},
					Age: time.Now().Add(-3 * time.Second),
				},
			},
			jobName:     "test-job",
			expectedLen: 2,
		},
		"returns empty slice when no matching pods": {
			pods: []resources.Pod{
				{
					Name: "pod-1",
					Labels: map[string]string{
						"job-name": "other-job",
					},
					Age: time.Now(),
				},
			},
			jobName:     "test-job",
			expectedLen: 0,
		},
		"returns pods sorted by age": {
			pods: []resources.Pod{
				{
					Name: "pod-2",
					Labels: map[string]string{
						"job-name": "test-job",
					},
					Age: time.Now().Add(-5 * time.Second),
				},
				{
					Name: "pod-1",
					Labels: map[string]string{
						"job-name": "test-job",
					},
					Age: time.Now().Add(-10 * time.Second),
				},
			},
			jobName:     "test-job",
			expectedLen: 2,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			result := getJobPodsFromDescribe(&testCase.pods, testCase.jobName)
			assert.Len(t, result, testCase.expectedLen)

			if testCase.expectedLen > 1 {
				for i := 1; i < len(result); i++ {
					assert.True(t, result[i-1].Age.Before(result[i].Age) || result[i-1].Age.Equal(result[i].Age))
				}
			}
		})
	}
}

func TestGetCreatedJobFromDescribe(t *testing.T) {
	testCases := map[string]struct {
		jobs     []resources.Job
		jobName  string
		expected bool
	}{
		"returns job when found": {
			jobs: []resources.Job{
				{Name: "job-1"},
				{Name: "test-job"},
				{Name: "job-2"},
			},
			jobName:  "test-job",
			expected: true,
		},
		"returns nil when not found": {
			jobs: []resources.Job{
				{Name: "job-1"},
				{Name: "job-2"},
			},
			jobName:  "test-job",
			expected: false,
		},
		"returns nil for empty slice": {
			jobs:     []resources.Job{},
			jobName:  "test-job",
			expected: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			result := getCreatedJobFromDescribe(&testCase.jobs, testCase.jobName)
			if testCase.expected {
				assert.NotNil(t, result)
				assert.Equal(t, testCase.jobName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
