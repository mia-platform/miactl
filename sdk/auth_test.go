package sdk

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
)

func TestGetProviders(t *testing.T) {
	testAppID := "my-app-id"

	tests := []struct {
		name              string
		requestAssertions requestAssertionFn
		responseBodyPath  string
		statusCode        int
		expectedOut       []Provider
		expectedErr       error
	}{
		{
			name: "returns providers",
			requestAssertions: func(t *testing.T, req *http.Request) {
				require.Equal(t, fmt.Sprintf("/apps/%s/providers/", testAppID), req.RequestURI)
			},
			responseBodyPath: "auth-providers.json",
			statusCode:       http.StatusOK,
			expectedOut: []Provider{{
				ID:   "my-provider-1",
				Type: "gitlab",
			}, {
				ID:   "my-provider-2",
				Type: "github",
			}},
		},
		{
			name: "error if returned providers payload is invalid",
			requestAssertions: func(t *testing.T, req *http.Request) {
				// FIXME: data race when this fails
				require.Equal(t, fmt.Sprintf("apps/%s/providers/", testAppID), req.RequestURI)
			},
			responseBodyPath: "auth-providers.json",
			statusCode:       http.StatusOK,
			expectedOut: []Provider{{
				ID:   "my-provider-1",
				Type: "gitlab",
			}, {
				ID:   "my-provider-2",
				Type: "github",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testData := readTestData(t, test.responseBodyPath)
			s := testCreateResponseServer(t, test.requestAssertions, testData, test.statusCode)
			defer s.Close()

			client := testCreateAuthClient(t, fmt.Sprintf("%s/", s.URL))

			providers, err := client.GetProviders(testAppID)
			require.Equal(t, test.expectedErr, err)
			require.Equal(t, test.expectedOut, providers)
		})
	}
}

func testCreateAuthClient(t *testing.T, url string) IAuth {
	t.Helper()

	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
	})
	require.NoError(t, err, "error creating client")

	return AuthClient{
		JSONClient: client,
	}
}
