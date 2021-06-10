package internal

import (
	"crypto/tls"
	"encoding/json"
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

	fileContent, err := ioutil.ReadFile(fmt.Sprintf("../testdata/%s", fileName))
	require.NoError(t, err)
	return string(fileContent)
}

type MockServerConfig struct {
	Endpoint       string
	Method         string
	QueryParams    map[string]interface{}
	RequestHeaders map[string]string
	RequestBody    interface{}
	Reply          interface{}
	ReplyStatus    int
}

type MockServerConfigs []MockServerConfig

type CertificatesConfig struct {
	CertPath string
	KeyPath  string
}

func MockServer(t *testing.T, cfgs MockServerConfigs, tlsCfg *CertificatesConfig) (*httptest.Server, error) {
	t.Helper()

	mux := http.NewServeMux()

	for _, c := range cfgs {
		expectedRequestBody, _ := json.Marshal(c.RequestBody)
		serverResponse, _ := json.Marshal(c.Reply)

		mux.HandleFunc(c.Endpoint, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			require.Equal(t, c.Endpoint, req.URL.Path, "correct endpoint")
			require.Equal(t, c.Method, req.Method)
			for key, expectedValue := range c.QueryParams {
				require.Equal(t, expectedValue, req.FormValue(key), "query param is the expected one")
			}
			for key, expectedValue := range c.RequestHeaders {
				require.Equal(t, expectedValue, req.Header.Get(key), "header value is the expected one")
			}

			// check body only if required (e.g. not needed for GET/DELETE method)
			if c.RequestBody != nil {
				rawRequestBody, _ := ioutil.ReadAll(req.Body)
				// remove the EOF/EOL (0xa byte) added by the ioutil.ReadAll to the request body bytes
				requestBody := rawRequestBody[:len(rawRequestBody)-1]

				require.Equal(t, expectedRequestBody, requestBody)
			}

			w.WriteHeader(c.ReplyStatus)
			if c.Reply != nil {
				w.Write(serverResponse)
			}
		}))
	}

	server := httptest.NewUnstartedServer(mux)

	if tlsCfg != nil {
		cert, err := tls.LoadX509KeyPair(tlsCfg.CertPath, tlsCfg.KeyPath)
		if err != nil {
			t.Fatalf("error loading TLS conf: %v", err)
			return nil, err
		}
		server.TLS = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server.StartTLS()
	} else {
		server.Start()
	}

	return server, nil
}
