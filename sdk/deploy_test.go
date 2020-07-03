package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestDeployGetHistory(t *testing.T) {
	projectsListResponseBody := readTestData(t, "projects.json")
	projectRequestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	historyRequestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.Equal(t, "/api/deploy/projects/mongo-id-2/deployment/", req.URL.Path)
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	t.Run("Error occurs when projectId does not exist in download list", func(t *testing.T) {
		s := testCreateResponseServer(t, projectRequestAssertions, projectsListResponseBody, 200)
		defer s.Close()
		client := testCreateDeployClient(t, s.URL)

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-NaN"})
		require.Nil(t, history)
		require.EqualError(t, err, fmt.Sprintf("%s: project-NaN", ErrProjectNotFound))
		require.True(t, errors.Is(err, ErrProjectNotFound))
	})

	t.Run("HTTP error occurs when downloading deploy history", func(t *testing.T) {
		historyResponseBody := `{"statusCode":500,"error":"InternalServerError","message":"some server error"}`
		responses := []response{
			{assertions: projectRequestAssertions, body: projectsListResponseBody, status: 200},
			{assertions: historyRequestAssertions, body: historyResponseBody, status: 500},
		}
		s := testCreateMultiResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClient(t, s.URL)

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, history)
		require.Error(t, err)
		require.True(t, errors.Is(err, jsonclient.ErrHTTP))
	})

	t.Run("Error on malformed history items (invalid DeployItem.ID)", func(t *testing.T) {
		historyResponseBody := readTestData(t, "deploy-history-invalid-payload.json")
		responses := []response{
			{assertions: projectRequestAssertions, body: projectsListResponseBody, status: 200},
			{assertions: historyRequestAssertions, body: historyResponseBody, status: 200},
		}
		s := testCreateMultiResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClient(t, s.URL)

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, history)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("%s: json: cannot unmarshal string into Go struct field DeployItem.id of type int", ErrGeneric))
		require.True(t, errors.Is(err, ErrGeneric))
	})

	t.Run("History download goes fine", func(t *testing.T) {
		historyResponseBody := readTestData(t, "deploy-history.json")
		responses := []response{
			{assertions: projectRequestAssertions, body: projectsListResponseBody, status: 200},
			{assertions: historyRequestAssertions, body: historyResponseBody, status: 200},
		}
		s := testCreateMultiResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClient(t, s.URL)

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, err)
		require.Equal(t, 3, len(history))

		deploy := history[0]
		commitDate, err := time.Parse(time.RFC3339, "2020-04-24T21:50:59.000+00:00")
		require.NoError(t, err)
		finishedAt, err := time.Parse(time.RFC3339, "2020-04-24T21:52:00.491Z")
		require.NoError(t, err)
		require.Equal(t, DeployItem{
			ID:     1234,
			Status: "success",
			Ref:    "v1.4.2",
			Commit: CommitInfo{
				URL:        "https://the-repo/123456789",
				AuthorName: "John Doe",
				CommitDate: commitDate,
				Hash:       "123456789",
			},
			User: DeployUser{
				Name: "John Doe",
			},
			DeployType:  "deploy_all",
			WebURL:      "https://the-repo/993344",
			Duration:    32.553293,
			FinishedAt:  finishedAt,
			Environment: "production",
		}, deploy)

		deploy = history[1]
		commitDate, err = time.Parse(time.RFC3339, "2020-04-24T21:04:13.000+00:00")
		require.NoError(t, err)
		finishedAt, err = time.Parse(time.RFC3339, "2020-04-24T21:05:08.633Z")
		require.NoError(t, err)
		require.Equal(t, DeployItem{
			ID:     1235,
			Status: "success",
			Ref:    "v1.4.1",
			Commit: CommitInfo{
				URL:        "https://the-repo/9876543",
				AuthorName: "Tim Applepie",
				CommitDate: commitDate,
				Hash:       "9876543",
			},
			User: DeployUser{
				Name: "Tim Applepie",
			},
			DeployType:  "deploy_all",
			WebURL:      "https://the-repo/443399",
			Duration:    30.759551,
			FinishedAt:  finishedAt,
			Environment: "production",
		}, deploy)

		deploy = history[2]
		commitDate, err = time.Parse(time.RFC3339, "2020-04-24T20:58:01.000+00:00")
		require.NoError(t, err)
		finishedAt, err = time.Parse(time.RFC3339, "2020-04-24T21:02:10.540Z")
		require.NoError(t, err)
		require.Equal(t, DeployItem{
			ID:     2414,
			Status: "failed",
			Ref:    "v1.4.0",
			Commit: CommitInfo{
				URL:        "https://the-repo/987123456",
				AuthorName: "F. Nietzsche",
				CommitDate: commitDate,
				Hash:       "987123456",
			},
			User: DeployUser{
				Name: "F. Nietzsche",
			},
			DeployType:  "deploy_all",
			WebURL:      "https://the-repo/334499",
			Duration:    32.671445,
			FinishedAt:  finishedAt,
			Environment: "development",
		}, deploy)
	})
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
