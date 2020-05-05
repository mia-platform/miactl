package sdk

// ProjectsMock is useful to be used to mock projects client
type ProjectsMock struct {
	Error    error
	Options  Options
	Projects Projects
}

// DeployMock is useful to be used to mock deploy client.
type DeployMock struct {
	Error    error
	AssertFn func(DeployHistoryQuery)
	History  []DeployItem
}

// MockClientError passes error to mia client mock
type MockClientError struct {
	ProjectsError error

	DeployError    error
	DeployAssertFn func(DeployHistoryQuery)
	DeployHistory  []DeployItem
}

// WrapperMockMiaClient creates a mock of mia client
func WrapperMockMiaClient(errors MockClientError) func(opts Options) (*MiaClient, error) {
	return func(opts Options) (*MiaClient, error) {
		return &MiaClient{
			Projects: &ProjectsMock{
				Error:   errors.ProjectsError,
				Options: opts,
			},
			Deploy: &DeployMock{
				Error:    errors.DeployError,
				AssertFn: errors.DeployAssertFn,
				History:  errors.DeployHistory,
			},
		}, nil
	}
}

// SetReturnError method set error to ProjectsMock
func (p *ProjectsMock) SetReturnError(err error) {
	p.Error = err
}

// SetReturnProjects method set projects to ProjectsMock. A project mock is
// returned by default calling Get function
func (p *ProjectsMock) SetReturnProjects(projects Projects) {
	p.Projects = projects
}

// Get method mock. It returns error or a list of projects
func (p ProjectsMock) Get() (Projects, error) {
	if p.Error != nil {
		return nil, p.Error
	}
	if p.Projects != nil {
		return p.Projects, nil
	}
	return defaultMockProjects, nil
}

var defaultMockProjects = Projects{
	Project{
		ID:                   "id1",
		Name:                 "Project 1",
		ConfigurationGitPath: "/git/path",
		Environments: []Environment{
			{
				Cluster: Cluster{
					Hostname: "cluster-hostname",
				},
				DisplayName: "development",
			},
		},
		Pipelines: Pipelines{
			Type: "pipeline-type",
		},
		ProjectID: "project-1",
	},
	Project{
		ID:                   "id2",
		Name:                 "Project 2",
		ConfigurationGitPath: "/git/path",
		Environments: []Environment{
			{
				Cluster: Cluster{
					Hostname: "cluster-hostname",
				},
				DisplayName: "development",
			},
		},
		ProjectID: "project-2",
	},
}

// GetHistory method mock. It returns error or a list of deploy items.
func (d DeployMock) GetHistory(query DeployHistoryQuery) ([]DeployItem, error) {
	if d.Error != nil {
		return nil, d.Error
	}

	if d.AssertFn != nil {
		d.AssertFn(query)
	}

	return d.History, nil
}
