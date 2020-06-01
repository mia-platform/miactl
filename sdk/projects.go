package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/davidebianchi/go-jsonclient"
)

// Cluster object, different for environment
type Cluster struct {
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
}

// Environment of the project
type Environment struct {
	DisplayName string  `json:"label"`
	EnvID       string  `json:"value"`
	Cluster     Cluster `json:"cluster"`
}

// Pipelines type supported by project
type Pipelines struct {
	Type string `json:"type"`
}

// Project define the mia-platform console project
type Project struct {
	ID                   string        `json:"_id"`
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
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
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
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
	}

	var project *Project
	for _, p := range projects {
		if p.ProjectID == projectID {
			project = &p
			break
		}
	}

	if project == nil {
		return nil, fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
	}
	return project, nil
}
