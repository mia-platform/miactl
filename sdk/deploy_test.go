package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	utils "github.com/mia-platform/miactl/sdk/internal"
	"github.com/stretchr/testify/require"
)

func TestDeployGetHistory(t *testing.T) {
	projectsListResponseBody := utils.ReadTestData(t, "projects.json")
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
		s := utils.CreateTestResponseServer(t, projectRequestAssertions, projectsListResponseBody, 200)
		defer s.Close()
		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-NaN"})
		require.Nil(t, history)
		require.EqualError(t, err, fmt.Sprintf("%s: project-NaN", ErrProjectNotFound))
		require.True(t, errors.Is(err, ErrProjectNotFound))
	})

	t.Run("HTTP error occurs when downloading deploy history", func(t *testing.T) {
		historyResponseBody := `{"statusCode":500,"error":"InternalServerError","message":"some server error"}`
		responses := utils.Responses{
			{Assertions: projectRequestAssertions, Body: projectsListResponseBody, Status: 200},
			{Assertions: historyRequestAssertions, Body: historyResponseBody, Status: 500},
		}
		s := utils.CreateMultiTestResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, history)
		require.Error(t, err)
		require.True(t, errors.Is(err, jsonclient.ErrHTTP))
	})

	t.Run("Error on malformed history items (invalid DeployItem.ID)", func(t *testing.T) {
		historyResponseBody := utils.ReadTestData(t, "deploy-history-invalid-payload.json")
		responses := utils.Responses{
			{Assertions: projectRequestAssertions, Body: projectsListResponseBody, Status: 200},
			{Assertions: historyRequestAssertions, Body: historyResponseBody, Status: 200},
		}
		s := utils.CreateMultiTestResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

		history, err := client.GetHistory(DeployHistoryQuery{ProjectID: "project-2"})
		require.Nil(t, history)
		require.Error(t, err)
		require.EqualError(t, err, fmt.Sprintf("%s: json: cannot unmarshal string into Go struct field DeployItem.id of type int", ErrGeneric))
		require.True(t, errors.Is(err, ErrGeneric))
	})

	t.Run("History download goes fine", func(t *testing.T) {
		historyResponseBody := utils.ReadTestData(t, "deploy-history.json")
		responses := utils.Responses{
			{Assertions: projectRequestAssertions, Body: projectsListResponseBody, Status: 200},
			{Assertions: historyRequestAssertions, Body: historyResponseBody, Status: 200},
		}
		s := utils.CreateMultiTestResponseServer(t, responses)
		defer s.Close()

		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

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

func TestTrigger(t *testing.T) {
	const (
		projectId   = "27ebd48c25a7"
		revision    = "master"
		environment = "development"
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
	)
	const expectedPipelineId = 458467
	expectedBearer := fmt.Sprintf("Bearer %s", apiToken)
	expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
	triggerEndpoint := fmt.Sprintf("api/deploy/projects/%s/trigger/pipeline/", projectId)

	t.Run("success - default behaviour", func(t *testing.T) {
		expectedResponse := DeployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}
		triggerAssertions := func(t *testing.T, req *http.Request) {
			t.Helper()

			expectedRequestBody := DeployRequest{
				Environment:             environment,
				Revision:                revision,
				DeployType:              SmartDeploy,
				ForceDeployWhenNoSemver: false,
			}

			require.True(t, strings.HasSuffix(req.URL.Path, triggerEndpoint))
			require.Equal(t, http.MethodPost, req.Method)
			require.Equal(t, expectedBearer, req.Header.Get("Authorization"))

			requestBody := DeployRequest{}
			bodyRequest, _ := io.ReadAll(req.Body)
			require.NoError(t, json.Unmarshal(bodyRequest, &requestBody))
			require.Equal(t, expectedRequestBody, requestBody)
		}
		// triggerResponse, err := json.Marshal(&expectedResponse)
		triggerResponse := fmt.Sprintf(`{"id":%d,"url":"%s"}`, expectedPipelineId, expectedPipelineURL)

		s := utils.CreateTestResponseServer(t, triggerAssertions, triggerResponse, http.StatusOK)
		defer s.Close()
		client := testCreateDeployClientToken(t, fmt.Sprintf("%s/", s.URL), apiToken)

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		require.Empty(t, err)
		require.Equal(t, expectedResponse, deployResponse)
	})

	t.Run("success - with deploy all strategy", func(t *testing.T) {
		expectedResponse := DeployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}
		triggerAssertions := func(t *testing.T, req *http.Request) {
			t.Helper()

			expectedRequestBody := DeployRequest{
				Environment:             environment,
				Revision:                revision,
				DeployType:              DeployAll,
				ForceDeployWhenNoSemver: true,
			}

			require.True(t, strings.HasSuffix(req.URL.Path, triggerEndpoint))
			require.Equal(t, http.MethodPost, req.Method)
			require.Equal(t, expectedBearer, req.Header.Get("Authorization"))

			requestBody := DeployRequest{}
			json.NewDecoder(req.Body).Decode(&requestBody)
			require.Equal(t, expectedRequestBody, requestBody)
		}
		triggerResponse := fmt.Sprintf(`{"id":%d,"url":"%s"}`, expectedPipelineId, expectedPipelineURL)

		s := utils.CreateTestResponseServer(t, triggerAssertions, triggerResponse, http.StatusOK)
		defer s.Close()
		client := testCreateDeployClientToken(t, fmt.Sprintf("%s/", s.URL), apiToken)

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
			DeployAll:   true,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		require.Empty(t, err)
		require.Equal(t, expectedResponse, deployResponse)
	})

	t.Run("failure", func(t *testing.T) {
		triggerAssertions := func(t *testing.T, req *http.Request) {
			t.Helper()

			expectedRequestBody := DeployRequest{
				Environment:             environment,
				Revision:                revision,
				DeployType:              SmartDeploy,
				ForceDeployWhenNoSemver: false,
			}

			require.True(t, strings.HasSuffix(req.URL.Path, triggerEndpoint))
			require.Equal(t, http.MethodPost, req.Method)
			require.Equal(t, expectedBearer, req.Header.Get("Authorization"))

			requestBody := DeployRequest{}
			json.NewDecoder(req.Body).Decode(&requestBody)
			require.Equal(t, expectedRequestBody, requestBody)
		}
		triggerResponse := "{}"

		s := utils.CreateTestResponseServer(t, triggerAssertions, triggerResponse, http.StatusBadRequest)
		defer s.Close()
		client := testCreateDeployClientToken(t, fmt.Sprintf("%s/", s.URL), apiToken)

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		base, _ := url.Parse(s.URL)
		path, _ := url.Parse(triggerEndpoint)
		require.EqualError(
			t,
			err,
			fmt.Sprintf("deploy error: POST %s: 400 - {}", base.ResolveReference(path)),
		)
		require.Empty(t, deployResponse)
	})
}
func TestGetDeployStatus(t *testing.T) {
	const (
		projectId   = "u543t8sdf34t5"
		pipelineId  = 32562
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
		environment = "preprod"
	)
	expectedBearer := fmt.Sprintf("Bearer %s", apiToken)

	t.Run("get status", func(t *testing.T) {
		expectedResponse := StatusResponse{
			PipelineId: pipelineId,
			Status:     Success,
		}
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		statusAssertions := func(t *testing.T, req *http.Request) {
			t.Helper()

			require.True(t, strings.HasSuffix(req.URL.Path, statusEndpoint))
			require.Equal(t, http.MethodGet, req.Method)
			require.Equal(t, expectedBearer, req.Header.Get("Authorization"))
			require.Equal(t, environment, req.FormValue("environment"))
		}
		statusResponse := fmt.Sprintf(`{"id":%d,"status":"%s"}`, pipelineId, Success)

		s := utils.CreateTestResponseServer(t, statusAssertions, statusResponse, http.StatusOK)
		defer s.Close()
		client := testCreateDeployClientToken(t, fmt.Sprintf("%s/", s.URL), apiToken)

		response, err := client.GetDeployStatus(projectId, pipelineId, environment)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, response)
	})

	t.Run("get status - error", func(t *testing.T) {
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		statusAssertions := func(t *testing.T, req *http.Request) {
			t.Helper()

			require.True(t, strings.HasSuffix(req.URL.Path, statusEndpoint))
			require.Equal(t, http.MethodGet, req.Method)
			require.Equal(t, expectedBearer, req.Header.Get("Authorization"))
			require.Equal(t, environment, req.FormValue("environment"))
		}
		statusResponse := fmt.Sprintf(`{"id":%d,"status":"%s"}`, pipelineId, Success)

		s := utils.CreateTestResponseServer(t, statusAssertions, statusResponse, http.StatusBadRequest)
		defer s.Close()
		client := testCreateDeployClientToken(t, fmt.Sprintf("%s/", s.URL), apiToken)

		response, err := client.GetDeployStatus(projectId, pipelineId, environment)
		require.Empty(t, response)

		require.Error(t, err)
		require.Contains(t, err.Error(), "status error:")
	})
}

func testCreateDeployClientCookie(t *testing.T, url string) IDeploy {
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

func testCreateDeployClientToken(t *testing.T, url, apiToken string) IDeploy {
	t.Helper()

	client, err := jsonclient.New(jsonclient.Options{
		BaseURL: url,
		Headers: jsonclient.Headers{
			"Authorization": fmt.Sprintf("Bearer %s", apiToken),
		},
	})
	require.NoError(t, err, "error creating client")

	return DeployClient{
		JSONClient: client,
	}
}
