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
	"testing"
	"time"

	"github.com/mia-platform/miactl/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestRowForCronJob(t *testing.T) {
	testCases := map[string]struct {
		cronjob     resources.CronJob
		expectedRow []string
	}{
		"basic cronjob": {
			cronjob: resources.CronJob{
				Name:         "cronjob-name",
				Suspend:      true,
				Active:       0,
				Schedule:     "* * * * *",
				Age:          time.Now().Add(-time.Hour * 24),
				LastSchedule: time.Now(),
			},
			expectedRow: []string{"cronjob-name", "* * * * *", "true", "0", "0s", "24h"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForCronJob(testCase.cronjob))
		})
	}
}

func TestRowForDeployment(t *testing.T) {
	testCases := map[string]struct {
		deployment  resources.Deployment
		expectedRow []string
	}{
		"basic deployment": {
			deployment: resources.Deployment{
				Name:      "deployment-name",
				Ready:     1,
				Replicas:  1,
				Available: 1,
				Age:       time.Now().Add(-time.Hour * 24),
			},
			expectedRow: []string{"deployment-name", "1/1", "1", "1", "24h"},
		},
		"missing ready and available": {
			deployment: resources.Deployment{
				Name:     "deployment-name",
				Replicas: 0,
				Age:      time.Now().Add(-time.Hour * 24),
			},
			expectedRow: []string{"deployment-name", "0/0", "0", "0", "24h"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForDeployment(testCase.deployment))
		})
	}
}

func TestRowForJob(t *testing.T) {
	testCases := map[string]struct {
		job         resources.Job
		expectedRow []string
	}{
		"basic job": {
			job: resources.Job{
				Name:           "job-name",
				Active:         0,
				Succeeded:      1,
				Failed:         0,
				Age:            time.Now().Add(-time.Hour * 24),
				StartTime:      time.Now().Add(-time.Second * 60),
				CompletionTime: time.Now(),
			},
			expectedRow: []string{"job-name", "1/1", "60s", "24h"},
		},
		"failed job": {
			job: resources.Job{
				Name:      "job-name",
				Failed:    1,
				Age:       time.Now().Add(-time.Hour * 24),
				StartTime: time.Now().Add(-time.Second * 60),
			},
			expectedRow: []string{"job-name", "0/1", "-", "24h"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForJob(testCase.job))
		})
	}
}

func TestRowForPod(t *testing.T) {
	testCases := map[string]struct {
		pod         resources.Pod
		expectedRow []string
	}{
		"basic pod": {
			pod: resources.Pod{
				Name:   "pod-name",
				Phase:  "running",
				Status: "ok",
				Age:    time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component", Version: "version"},
				},
				Containers: []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					Status       string `json:"status"`
				}{
					{
						Name:         "container-name",
						Ready:        true,
						RestartCount: 0,
						Status:       "running",
					},
				},
			},
			expectedRow: []string{"Ok", "pod-name", "component:version", "1/1", "Running", "0", "0s"},
		},
		"pod without component": {
			pod: resources.Pod{
				Age: time.Now(),
			},
			expectedRow: []string{"", "", "-", "0/0", "", "0", "0s"},
		},
		"pod without component version": {
			pod: resources.Pod{
				Age: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component"},
				},
			},
			expectedRow: []string{"", "", "component", "0/0", "", "0", "0s"},
		},
		"pod without component name": {
			pod: resources.Pod{
				Age: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Version: "version"},
				},
			},
			expectedRow: []string{"", "", "-", "0/0", "", "0", "0s"},
		},
		"pod without multiple components": {
			pod: resources.Pod{
				Age: time.Now(),
				Component: []struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{
					{Name: "component", Version: "version"},
					{Name: "component"},
				},
			},
			expectedRow: []string{"", "", "component:version, component", "0/0", "", "0", "0s"},
		},
		"pod with multiple containers": {
			pod: resources.Pod{
				Age: time.Now(),
				Containers: []struct {
					Name         string `json:"name"`
					Ready        bool   `json:"ready"`
					RestartCount int    `json:"restartCount"`
					Status       string `json:"status"`
				}{
					{
						Name:         "container-name",
						Ready:        true,
						RestartCount: 3,
						Status:       "running",
					},
					{
						Name:         "container-name2",
						Ready:        false,
						RestartCount: 1,
						Status:       "running",
					},
				},
			},
			expectedRow: []string{"", "", "-", "1/2", "", "4", "0s"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForPod(testCase.pod))
		})
	}
}

func TestRowForService(t *testing.T) {
	testCases := map[string]struct {
		service     resources.Service
		expectedRow []string
	}{
		"basic service": {
			service: resources.Service{
				Name:      "service-name",
				Type:      "ClusterIP",
				ClusterIP: "127.0.0.1",
				Ports: []resources.Port{
					{
						Name:       "port-name",
						Port:       8000,
						Protocol:   "TCP",
						TargetPort: "8000",
					},
				},
				Age: time.Now().Add(-time.Hour * 24),
			},
			expectedRow: []string{"service-name", "ClusterIP", "127.0.0.1", "8000/TCP", "24h"},
		},
		"missing cluster ip": {
			service: resources.Service{
				Name: "service-name",
				Type: "ClusterIP",
				Ports: []resources.Port{
					{
						Name:       "port-name",
						Port:       8000,
						Protocol:   "TCP",
						TargetPort: "8000",
					},
				},
				Age: time.Now().Add(-time.Hour * 24),
			},
			expectedRow: []string{"service-name", "ClusterIP", "<none>", "8000/TCP", "24h"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRow, rowForService(testCase.service))
		})
	}
}
