package internal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

type AssertionFn func(t *testing.T, req *http.Request)

type Response struct {
	Assertions AssertionFn
	Body       string
	Status     int
}
type Responses []Response

func CreateTestClient(t *testing.T, url string) *jsonclient.Client {
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

func CreateTestResponseServer(t *testing.T, assertions AssertionFn, responseBody string, statusCode int) *httptest.Server {
	t.Helper()
	responses := []Response{
		{Assertions: assertions, Body: responseBody, Status: statusCode},
	}
	return CreateMultiTestResponseServer(t, responses)
}

func CreateMultiTestResponseServer(t *testing.T, responses Responses) *httptest.Server {
	t.Helper()
	var usage int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if usage >= len(responses) {
			t.Fatalf("Unexpected HTTP request, provided %d handler, received call #%d.", len(responses), usage+1)
		}

		response := responses[usage]
		usage++
		if response.Assertions != nil {
			response.Assertions(t, req)
		}

		w.WriteHeader(response.Status)
		var responseBytes []byte
		if response.Body != "" {
			responseBytes = []byte(response.Body)
		}
		w.Write(responseBytes)
	}))
}

func ReadTestData(t *testing.T, fileName string) string {
	t.Helper()

	fileContent, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", fileName))
	require.NoError(t, err)
	return string(fileContent)
}
