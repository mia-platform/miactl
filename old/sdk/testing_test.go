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

package sdk

import (
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/old/sdk/deploy"
	"github.com/stretchr/testify/require"
)

func TestWrapperMockMiaClient(t *testing.T) {
	setupClient := func(t *testing.T) *ProjectsMock {
		wrapperClient := WrapperMockMiaClient(MockClientError{})

		opts := Options{
			APIKey: "apiKey",
		}
		miaClient, err := wrapperClient(opts)
		require.NoError(t, err)

		projectClient, ok := miaClient.Projects.(ProjectsMock)
		if ok {
			t.Fail()
		}

		return &projectClient
	}

	t.Run("returns mia client mock", func(t *testing.T) {
		prjErr := fmt.Errorf("error project")
		wrapperClient := WrapperMockMiaClient(MockClientError{
			ProjectsError: prjErr,
		})

		opts := Options{
			APIKey: "apiKey",
		}
		miaClient, err := wrapperClient(opts)
		require.NoError(t, err)

		require.Equal(t, &MiaClient{
			Projects: &ProjectsMock{
				Error:   prjErr,
				Options: opts,
			},
			Deploy: &DeployMock{
				Error:    nil,
				AssertFn: nil,
				History:  nil,
			},
		}, miaClient)
	})

	t.Run("set error on mock project client and returns with Get method", func(t *testing.T) {
		prjClient := setupClient(t)

		prjErr := fmt.Errorf("error project")
		prjClient.SetReturnError(prjErr)

		require.Equal(t, &ProjectsMock{
			Error: prjErr,
		}, prjClient)

		retProjects, err := prjClient.Get()
		require.Nil(t, retProjects)
		require.EqualError(t, err, prjErr.Error())
	})

	t.Run("set projects on mock project client and returns it with Get method", func(t *testing.T) {
		prjClient := setupClient(t)

		projects := deploy.Projects{
			deploy.Project{
				ID:                   "id-prova",
				Name:                 "Project 1",
				ConfigurationGitPath: "/git/path",
				Environments: []deploy.Environment{
					{
						Cluster: deploy.Cluster{
							Hostname: "cluster-hostname",
						},
						DisplayName: "development",
					},
				},
				ProjectID: "project-1",
			},
		}
		prjClient.SetReturnProjects(projects)

		require.Equal(t, &ProjectsMock{
			Projects: projects,
		}, prjClient)

		retProjects, err := prjClient.Get()
		require.NoError(t, err)
		require.Equal(t, projects, retProjects)
	})

	t.Run("set projects on mock project client", func(t *testing.T) {
		prjClient := setupClient(t)

		retProjects, err := prjClient.Get()
		require.NoError(t, err)
		require.Equal(t, defaultMockProjects, retProjects)
	})
}
