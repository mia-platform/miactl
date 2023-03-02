package request

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

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
		authFn: mockSuccessAuthFn,
	}
	actualReq := RequestBuilder(*opts, mockSuccessAuthFn)
	require.Equal(t, expectedReq.url, actualReq.url)
	require.Equal(t, expectedReq.token, actualReq.token)
	require.NotNil(t, actualReq.authFn)
	require.NotNil(t, actualReq.client)
}

func mockSuccessAuthFn(url string) (string, error) {
	return "new_token", nil
}
