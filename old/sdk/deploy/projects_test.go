package deploy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"
	utils "github.com/mia-platform/miactl/old/sdk/internal"
	"github.com/stretchr/testify/require"
)

func TestProjectsGet(t *testing.T) {
	projectsListResponseBody := utils.ReadTestData(t, "projects.json")
	requestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	t.Run("correctly returns projects", func(t *testing.T) {
		expectedProjects := Projects{
			Project{
				ID:                   "mongo-id-1",
				Name:                 "Project 1",
				ConfigurationGitPath: "/clients/path",
				ProjectID:            "project-1",
				Environments: []Environment{
					{
						Cluster: Cluster{
							Hostname:  "127.0.0.1",
							Namespace: "project-1-dev",
						},
						DisplayName: "Development",
						EnvID:       "development",
					},
				},
				Pipelines: Pipelines{
					Type: "gitlab",
				},
			},
			Project{
				ID:                   "mongo-id-2",
				Name:                 "Project 2",
				ConfigurationGitPath: "/clients/path/configuration",
				ProjectID:            "project-2",
				Environments: []Environment{
					{
						Cluster: Cluster{
							Hostname:  "127.0.0.1",
							Namespace: "project-2-dev",
						},
						DisplayName: "Development",
						EnvID:       "development",
					},
					{
						Cluster: Cluster{
							Hostname:  "127.0.0.1",
							Namespace: "project-2",
						},
						DisplayName: "Production",
						EnvID:       "production",
					},
				},
			},
		}

		s := utils.CreateTestResponseServer(t, requestAssertions, projectsListResponseBody, 200)
		client := testCreateProjectClient(t, fmt.Sprintf("%s/", s.URL))

		projects, err := client.Get()
		require.NoError(t, err)
		require.Equal(t, expectedProjects, projects)
	})

	t.Run("throws when server respond with 401", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := utils.CreateTestResponseServer(t, requestAssertions, responseBody, 401)
		client := testCreateProjectClient(t, fmt.Sprintf("%s/", s.URL))

		projects, err := client.Get()
		require.Nil(t, projects)
		require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/: 401 - %s", s.URL, responseBody))
		require.True(t, errors.Is(err, sdkErrors.ErrHTTP))
	})

	t.Run("throws if response body is not as expected", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := utils.CreateTestResponseServer(t, requestAssertions, responseBody, 200)
		client := testCreateProjectClient(t, fmt.Sprintf("%s/", s.URL))

		projects, err := client.Get()
		require.Nil(t, projects)
		require.Error(t, err)
		require.True(t, errors.Is(err, sdkErrors.ErrGeneric))
	})
}

func TestGetProjectByID(t *testing.T) {
	projectsListResponseBody := utils.ReadTestData(t, "projects.json")
	projectRequestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	t.Run("Unauthorized error occurs during projectId fetch", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := utils.CreateTestResponseServer(t, projectRequestAssertions, responseBody, 401)
		defer s.Close()

		client := utils.CreateTestClient(t, fmt.Sprintf("%s/", s.URL))
		project, err := getProjectByID(client, "project1")
		require.Nil(t, project)
		require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/: 401 - %s", s.URL, responseBody))
		require.True(t, errors.Is(err, sdkErrors.ErrHTTP))
	})

	t.Run("Generic error occurs during projectId fetch (malformed data, _id should be a string)", func(t *testing.T) {
		responseBody := utils.ReadTestData(t, "projects-invalid-payload.json")
		s := utils.CreateTestResponseServer(t, projectRequestAssertions, responseBody, 200)
		defer s.Close()

		client := utils.CreateTestClient(t, fmt.Sprintf("%s/", s.URL))
		project, err := getProjectByID(client, "project1")
		require.Nil(t, project)
		require.EqualError(t, err, fmt.Sprintf("%s: json: cannot unmarshal number into Go struct field Project._id of type string", sdkErrors.ErrGeneric))
		require.True(t, errors.Is(err, sdkErrors.ErrGeneric))
	})

	t.Run("Error projectID not found", func(t *testing.T) {
		s := utils.CreateTestResponseServer(t, projectRequestAssertions, projectsListResponseBody, 200)
		defer s.Close()

		client := utils.CreateTestClient(t, fmt.Sprintf("%s/", s.URL))
		project, err := getProjectByID(client, "project1")
		require.Nil(t, project)
		require.EqualError(t, err, fmt.Sprintf("%s: project1", sdkErrors.ErrProjectNotFound))
		require.True(t, errors.Is(err, sdkErrors.ErrProjectNotFound))
	})

	t.Run("Returns desired project", func(t *testing.T) {
		s := utils.CreateTestResponseServer(t, projectRequestAssertions, projectsListResponseBody, 200)
		defer s.Close()

		client := utils.CreateTestClient(t, fmt.Sprintf("%s/", s.URL))
		project, err := getProjectByID(client, "project-2")
		require.NoError(t, err)
		require.Equal(t, &Project{
			ID:                   "mongo-id-2",
			Name:                 "Project 2",
			ConfigurationGitPath: "/clients/path/configuration",
			ProjectID:            "project-2",
			Environments: []Environment{{
				EnvID:       "development",
				DisplayName: "Development",
				Cluster: Cluster{
					Hostname:  "127.0.0.1",
					Namespace: "project-2-dev",
				},
			}, {
				EnvID:       "production",
				DisplayName: "Production",
				Cluster: Cluster{
					Hostname:  "127.0.0.1",
					Namespace: "project-2",
				},
			}},
		},
			project)
	})
}

func testCreateProjectClient(t *testing.T, url string) IProjects {
	t.Helper()
	return ProjectsClient{
		JSONClient: utils.CreateTestClient(t, url),
	}
}
