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
	projectRequestAssertions := func(t *testing.T, req *http.Request) {
		t.Helper()

		require.True(t, strings.HasSuffix(req.URL.Path, "/projects/"))
		require.Equal(t, http.MethodGet, req.Method)
		cookieSid, err := req.Cookie("sid")
		require.NoError(t, err)
		require.Equal(t, &http.Cookie{Name: "sid", Value: "my-random-sid"}, cookieSid)
	}

	historyRequestAsserions := func(t *testing.T, req *http.Request) {
		t.Helper()

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
			{assertions: historyRequestAsserions, body: historyResponseBody, status: 500},
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
		historyResponseBody := `[{"id":"abcde","status":"success","ref":"v1.4.2","commit":{"url":"https://the-repo/123456789","authorName":"John Doe","committedDate":"2020-04-24T21:50:59.000+00:00","sha":"123456789"},"user":{"name":"John Doe"},"deployType":"deploy_all","webUrl":"https://the-repo/993344","duration":32.553293,"finishedAt":"2020-04-24T21:52:00.491Z","env":"production"}]`
		responses := []response{
			{assertions: projectRequestAssertions, body: projectsListResponseBody, status: 200},
			{assertions: historyRequestAsserions, body: historyResponseBody, status: 200},
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
		historyResponseBody := `[{"id":1234,"status":"success","ref":"v1.4.2","commit":{"url":"https://the-repo/123456789","authorName":"John Doe","committedDate":"2020-04-24T21:50:59.000+00:00","sha":"123456789"},"user":{"name":"John Doe"},"deployType":"deploy_all","webUrl":"https://the-repo/993344","duration":32.553293,"finishedAt":"2020-04-24T21:52:00.491Z","env":"production"},{"id":1235,"status":"success","ref":"v1.4.1","commit":{"url":"https://the-repo/9876543","authorName":"Tim Applepie","committedDate":"2020-04-24T21:04:13.000+00:00","sha":"9876543"},"user":{"name":"Tim Applepie"},"deployType":"deploy_all","webUrl":"https://the-repo/443399","duration":30.759551,"finishedAt":"2020-04-24T21:05:08.633Z","env":"production"},{"id":2414,"status":"failed","ref":"v1.4.0","commit":{"url":"https://the-repo/987123456","authorName":"F. Nietzsche","committedDate":"2020-04-24T20:58:01.000+00:00","sha":"987123456"},"user":{"name":"F. Nietzsche"},"deployType":"deploy_all","webUrl":"https://the-repo/334499","duration":32.671445,"finishedAt":"2020-04-24T21:02:10.540Z","env":"development"}]`
		responses := []response{
			{assertions: projectRequestAssertions, body: projectsListResponseBody, status: 200},
			{assertions: historyRequestAsserions, body: historyResponseBody, status: 200},
		}
		s := testCreateMultiResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClient(t, s.URL)

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, err)
		require.Equal(t, 3, len(history))
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
