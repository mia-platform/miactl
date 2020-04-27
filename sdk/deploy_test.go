package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestDeployGetHistory(t *testing.T) {
	requestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	t.Run("Error occurs during projectId fetch", func(t *testing.T) {
		responseBody := `{"statusCode":401,"error":"Unauthorized","message":"Unauthorized"}`
		s := testCreateResponseServer(t, requestAssertions, responseBody, 401)
		client := testCreateDeployClient(t, s.URL)

		projects, err := client.GetHistory("project1")
		require.Nil(t, projects)
		require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/: 401 - %s", s.URL, responseBody))
		require.True(t, errors.Is(err, ErrHTTP))
	})

	t.Run("Error occurs when projectId does not exist in download list", func(t *testing.T) {
		projectsResponseBody := `[{"_id":"mongo-id-1","name":"Project 1","configurationGitPath":"/clients/path","projectId":"project-1","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-1-dev"}}],"pipelines":{"type":"gitlab"}},{"_id":"mongo-id-2","name":"Project 2","configurationGitPath":"/clients/path/configuration","projectId":"project-2","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-2-dev"}},{"label":"Production","value":"production","cluster":{"hostname":"127.0.0.1","namespace":"project-2"}}]}]`

		s := testCreateResponseServer(t, requestAssertions, projectsResponseBody, 200)
		defer s.Close()
		client := testCreateDeployClient(t, s.URL)

		projects, err := client.GetHistory("project-NaN")
		require.Nil(t, projects)
		require.EqualError(t, err, fmt.Sprintf("%s: project-NaN", ErrProjectNotFound))
		require.True(t, errors.Is(err, ErrProjectNotFound))
	})

	t.Run("Error occurs when downloading deploy history", func(t *testing.T) {
		projectsResponseBody := `[{"_id":"mongo-id-1","name":"Project 1","configurationGitPath":"/clients/path","projectId":"project-1","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-1-dev"}}],"pipelines":{"type":"gitlab"}},{"_id":"mongo-id-2","name":"Project 2","configurationGitPath":"/clients/path/configuration","projectId":"project-2","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-2-dev"}},{"label":"Production","value":"production","cluster":{"hostname":"127.0.0.1","namespace":"project-2"}}]}]`
		historyResponseBody := `{"message":"some server error"}`
		responses := []response{
			{assertions: requestAssertions, body: projectsResponseBody, status: 200},
			{assertions: requestAssertions, body: historyResponseBody, status: 500},
		}
		s := testCreateMultiResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClient(t, s.URL)

		projects, err := client.GetHistory("project-2")
		require.Nil(t, projects)
		require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/mongo-id-2/deployment/?page=1&per_page=25&sort=desc: 500 - %s", s.URL, historyResponseBody))
		require.True(t, errors.Is(err, ErrHTTP))
	})

	// t.Run("Downloads occurs when downloading deploy history", func(t *testing.T) {
	// 	projectsResponseBody := `[{"_id":"mongo-id-1","name":"Project 1","configurationGitPath":"/clients/path","projectId":"project-1","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-1-dev"}}],"pipelines":{"type":"gitlab"}},{"_id":"mongo-id-2","name":"Project 2","configurationGitPath":"/clients/path/configuration","projectId":"project-2","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-2-dev"}},{"label":"Production","value":"production","cluster":{"hostname":"127.0.0.1","namespace":"project-2"}}]}]`
	// 	historyResponseBody := `{"message":"some server error"}`
	// 	responses := []response{
	// 		{assertions: requestAssertions, body: projectsResponseBody, status: 200},
	// 		{assertions: requestAssertions, body: historyResponseBody, status: 500},
	// 	}
	// 	s := testCreateMultiResponseServer(t, responses)
	// 	client := testCreateDeployClient(t, s.URL)

	// 	projects, err := client.GetHistory("project-2")
	// 	require.Nil(t, projects)
	// 	require.EqualError(t, err, fmt.Sprintf("GET %s/api/backend/projects/mongo-id-2/deployment/?page=1&per_page=25&sort=desc: 500 - %s", s.URL, historyResponseBody))
	// 	require.True(t, errors.Is(err, ErrHTTP))
	// })
}

func testCreateDeployClient(t *testing.T, url string) IDeploy {
	t.Helper()

	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
		Headers: jsonclient.Headers{
			"cookie": "sid=my-random-sid",
		},
	})
	require.NoError(t, err, "error creating client")

	return DeployClient{
		JSONClient: client,
	}
}
