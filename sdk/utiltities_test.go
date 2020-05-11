package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

const projectsListResponseBody = `[{"_id":"mongo-id-1","name":"Project 1","configurationGitPath":"/clients/path","projectId":"project-1","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-1-dev"}}],"pipelines":{"type":"gitlab"}},{"_id":"mongo-id-2","name":"Project 2","configurationGitPath":"/clients/path/configuration","projectId":"project-2","environments":[{"label":"Development","value":"development","cluster":{"hostname":"127.0.0.1","namespace":"project-2-dev"}},{"label":"Production","value":"production","cluster":{"hostname":"127.0.0.1","namespace":"project-2"}}]}]`

type assertionFn func(t *testing.T, req *http.Request)

type response struct {
	assertions assertionFn
	body       string
	status     int
}
type responses []response

func testCreateClient(t *testing.T, url string) *jsonclient.Client {
	t.Helper()
	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
		Headers: jsonclient.Headers{
			"cookie": "sid=my-random-sid",
		},
	})
	require.NoError(t, err, "error creating client")
	return client
}

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
		if usage >= len(responses) {
			t.Fatalf("Unexpected HTTP request, provided %d handler, received call #%d.", len(responses), usage+1)
		}

		response := responses[usage]
		usage++
		if response.assertions != nil {
			response.assertions(t, req)
		}

		w.WriteHeader(response.status)
		var responseBytes []byte
		if response.body != "" {
			responseBytes = []byte(response.body)
		}
		w.Write(responseBytes)
	}))
}
