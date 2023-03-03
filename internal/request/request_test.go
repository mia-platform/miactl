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
		APIBaseURL: "url",
		APIToken:   "token",
	}
	expectedReq := &Request{
		url:    "url",
		token:  "token",
		client: &http.Client{},
		authFn: mockSuccessAuthFn,
	}
	actualReq := RequestBuilder(*opts, mockSuccessAuthFn)
	require.Equal(t, expectedReq.url, actualReq.url)
	require.Equal(t, expectedReq.token, actualReq.token)
	require.NotNil(t, actualReq.authFn)
}

func TestExecute(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://testurl.io/testget",
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
	validReq := &Request{
		url:    "https://testurl.io/testget",
		token:  "valid_token",
		client: &http.Client{},
		authFn: mockSuccessAuthFn,
	}
	resp, err := validReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request with expired token
	expReq := &Request{
		url:    "https://testurl.io/testget",
		token:  "expired_token",
		client: &http.Client{},
		authFn: mockSuccessAuthFn,
	}
	resp, err = expReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request without token
	unauthReq := &Request{
		url:    "https://testurl.io/testget",
		client: &http.Client{},
		authFn: mockSuccessAuthFn,
	}
	resp, err = unauthReq.Execute()
	require.Nil(t, err)
	require.Equal(t, "200", resp.Status)

	// Test request without token
	failAuthReq := &Request{
		url:    "https://testurl.io/testget",
		client: &http.Client{},
		authFn: mockFailAuthFn,
	}
	resp, err = failAuthReq.Execute()
	require.Equal(t, "401", resp.Status)
	require.Equal(t, "error in authentication flow: authentication failed", err.Error())
}

func mockSuccessAuthFn(url string) (string, error) {
	return "valid_token", nil
}

func mockFailAuthFn(url string) (string, error) {
	return "", fmt.Errorf("authentication failed")
}
