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

package deploy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"
)

// IProjects expose the projects client interface
type IProjects interface {
	Get() (Projects, error)
}

// HistoryQuery wraps query filters for project deployments.
type HistoryQuery struct {
	ProjectID string
}

// Cluster object, different for environment
type Cluster struct {
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
}

// Environment of the project
type Environment struct {
	DisplayName string  `json:"label"` //nolint:tagliatelle
	EnvID       string  `json:"value"` //nolint:tagliatelle
	Cluster     Cluster `json:"cluster"`
}

// Pipelines type supported by project
type Pipelines struct {
	Type string `json:"type"`
}

// Project define the mia-platform console project
type Project struct {
	ID                   string        `json:"_id"` //nolint:tagliatelle
	Name                 string        `json:"name"`
	ConfigurationGitPath string        `json:"configurationGitPath"`
	Environments         []Environment `json:"environments"`
	ProjectID            string        `json:"projectId"`
	Pipelines            Pipelines     `json:"pipelines"`
}

// Projects is a list of project
type Projects []Project

// ProjectsClient is the console implementations of the IProjects interface
type ProjectsClient struct {
	JSONClient *jsonclient.Client
}

// Get method to fetch the console projects
func (p ProjectsClient) Get() (Projects, error) {
	req, err := p.JSONClient.NewRequest(http.MethodGet, "api/backend/projects/", nil)
	if err != nil {
		return nil, err
	}

	projects := Projects{}
	var httpErr *jsonclient.HTTPError
	_, err = p.JSONClient.Do(req, &projects)
	if err != nil {
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrGeneric, err)
	}

	return projects, nil
}

func getProjectByID(client *jsonclient.Client, projectID string) (*Project, error) {
	req, err := client.NewRequest(http.MethodGet, "api/backend/projects/", nil)
	if err != nil {
		return nil, err
	}

	var projects Projects
	if _, err := client.Do(req, &projects); err != nil {
		var httpErr *jsonclient.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrGeneric, err)
	}

	var project *Project
	for _, p := range projects {
		p := p
		if p.ProjectID == projectID {
			project = &p
			break
		}
	}

	if project == nil {
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrProjectNotFound, projectID)
	}
	return project, nil
}
