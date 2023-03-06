package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/mia-platform/miactl/old/sdk"
	"github.com/stretchr/testify/require"
)

const testUrl = "https://testurl.io/testget"

var (
	testToken string
	client    = &http.Client{}
)

func TestWithBody(t *testing.T) {
	req := &Request{}
	values := map[string]string{"key": "value"}
	jsonValues, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	body := bytes.NewBuffer(jsonValues)
	wrappedBody := io.NopCloser(body)
	req.WithBody(wrappedBody)
	require.Equal(t, wrappedBody, req.body)
}

func TestGet(t *testing.T) {
	req := &Request{}
	req.Get()
	require.Equal(t, "GET", req.method)
}

func TestPost(t *testing.T) {
	req := &Request{}
	values := map[string]string{"key": "value"}
	jsonValues, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	body := bytes.NewBuffer(jsonValues)
	wrappedBody := io.NopCloser(body)
	req.Post(wrappedBody)
	require.Equal(t, "POST", req.method)
	require.Equal(t, wrappedBody, req.body)
}

func TestRequestBuilder(t *testing.T) {
	opts := &sdk.Options{
		APIBaseURL: testUrl,
	}
	expectedReq := &Request{
		url:    testUrl,
		client: client,
		authFn: mockValidToken,
	}
	actualReq := RequestBuilder(*opts, mockValidToken)
	require.Equal(t, expectedReq.url, actualReq.url)
	require.NotNil(t, actualReq.authFn)
}

func TestExecute(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", testUrl,
		func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error
			if req.Header.Get("Authorization") != "Bearer valid_token" {
				resp, err = httpmock.NewJsonResponse(401, map[string]interface{}{
					"authorized": "false",
				})
			} else {
				resp, err = httpmock.NewJsonResponse(200, map[string]interface{}{
					"authorized": "true",
				})
			}
			return resp, err
		},
	)

	// Test request with valid token
	testToken = ""
	validReq := &Request{
		url:    testUrl,
		client: client,
		authFn: mockValidToken,
	}
	resp, err := validReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request with expired token
	testToken = ""
	expReq := &Request{
		url:    testUrl,
		client: client,
		authFn: mockExpiredToken,
	}
	resp, err = expReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test auth error
	testToken = ""
	failAuthReq := &Request{
		url:    testUrl,
		client: client,
		authFn: mockFailAuth,
	}
	resp, err = failAuthReq.Execute()
	require.Nil(t, resp)
	require.Equal(t, "error retrieving token: authentication failed", err.Error())

	// Test token refresh error
	testToken = ""
	failRefreshReq := &Request{
		url:    testUrl,
		client: client,
		authFn: mockFailRefresh,
	}
	resp, err = failRefreshReq.Execute()
	require.Equal(t, "401", resp.Status)
	require.Equal(t, "error refreshing token: authentication failed", err.Error())
}

func mockValidToken(url string) (string, error) {
	return "valid_token", nil
}

func mockExpiredToken(url string) (string, error) {
	if testToken == "" {
		testToken = "expired_token"
	} else {
		testToken = "valid_token"
	}
	return testToken, nil
}

func mockFailAuth(url string) (string, error) {
	return "", fmt.Errorf("authentication failed")
}

func mockFailRefresh(url string) (string, error) {
	if testToken == "" {
		testToken = "expired_token"
		return testToken, nil
	}
	return "", fmt.Errorf("authentication failed")
}
