package sdk

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
