package resources

type Cluster struct {
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
}

type Environment struct {
	DisplayName string  `json:"label"` //nolint:tagliatelle
	EnvID       string  `json:"value"` //nolint:tagliatelle
	Cluster     Cluster `json:"cluster"`
}
type Pipelines struct {
	Type string `json:"type"`
}

type Project struct {
	ID                   string        `json:"_id"` //nolint:tagliatelle
	Name                 string        `json:"name"`
	ConfigurationGitPath string        `json:"configurationGitPath"`
	Environments         []Environment `json:"environments"`
	ProjectID            string        `json:"projectId"`
	Pipelines            Pipelines     `json:"pipelines"`
	TenantID             string        `json:"tenantId"`
}
