package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestProjectsGet(t *testing.T) {
	requestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	t.Run("correctly returns projects", func(t *testing.T) {
		responseBody := `[{"_id":"mongo-id-1","name":"Project 1","configurationGitPath":"/clients/path","projectId":"project-1","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-1-dev"}}],"pipelines":{"type":"gitlab"}},{"_id":"mongo-id-2","name":"Project 2","configurationGitPath":"/clients/path/configuration","projectId":"project-2","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-2-dev"}},{"label":"Production","value":"production","cluster":{"hostname":"127.0.0.1","namespace":"project-2"}}]}]`
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

		s := testCreateResponseServer(t, requestAssertions, responseBody, 200)
		client := testCreateProjectClient(t, s.URL)

		projects, err := client.Get()
		require.NoError(t, err)
		require.Equal(t, expectedProjects, projects)
	})

	t.Run("throws when server respond with 401", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := testCreateResponseServer(t, requestAssertions, responseBody, 401)
		client := testCreateProjectClient(t, s.URL)

		projects, err := client.Get()
		require.Nil(t, projects)
		require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/: 401 - %s", s.URL, responseBody))
		require.True(t, errors.Is(err, ErrHTTP))
	})

	t.Run("throws if response body is not as expected", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := testCreateResponseServer(t, requestAssertions, responseBody, 200)
		client := testCreateProjectClient(t, s.URL)

		projects, err := client.Get()
		require.Nil(t, projects)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrGeneric))
	})
}

func testCreateProjectClient(t *testing.T, url string) IProjects {
	t.Helper()

	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
		Headers: jsonclient.Headers{
			"cookie": "sid=my-random-sid",
		},
	})
	require.NoError(t, err, "error creating client")

	return ProjectsClient{
		JSONClient: client,
	}
}

type assertionFn func(t *testing.T, req *http.Request)

type response struct {
	assertions assertionFn
	body       string
	status     int
}
type responses []response

func testCreateResponseServer(t *testing.T, assertions assertionFn, responseBody string, statusCode int) *httptest.Server {
	t.Helper()
	responses := []response{
		{assertions: assertions, body: responseBody, status: statusCode},
	}
	return testCreateMultiResponseServer(t, responses)
}

func testCreateMultiResponseServer(t *testing.T, responses responses) *httptest.Server {
	t.Helper()

	var usage int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Invoked", usage, req.URL.String())
		response := responses[usage]
		usage++
		if response.assertions != nil {
			response.assertions(t, req)
		}

		w.WriteHeader(response.status)
		if response.body != "" {
			w.Write([]byte(response.body))
			return
		}
		w.Write(nil)
	}))
}
