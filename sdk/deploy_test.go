package sdk

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

type sleeperMock struct {
	CallCount int
}

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
		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

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

		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

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

		client := testCreateDeployClientCookie(t, fmt.Sprintf("%s/", s.URL))

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
	expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
	triggerEndpoint := fmt.Sprintf("api/deploy/projects/%s/trigger/pipeline/", projectId)

	client := testCreateDeployClientToken(t, baseURL, apiToken)

	t.Run("success - default behaviour", func(t *testing.T) {
		defer gock.Off()

		expectedResponse := DeployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}

		gock.New(baseURL).
			Post(triggerEndpoint).
			MatchHeader("Authorization", fmt.Sprintf("Bearer %s", apiToken)).
			MatchType("json").
			JSON(map[string]interface{}{
				"environment":             environment,
				"revision":                revision,
				"deployType":              SmartDeploy,
				"forceDeployWhenNoSemver": false,
			}).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		require.Empty(t, err)
		require.Equal(t, expectedResponse, deployResponse)

		require.True(t, gock.IsDone())
	})

	t.Run("success - with deploy all strategy", func(t *testing.T) {
		defer gock.Off()

		const expectedPipelineId = 458467
		expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
		expectedResponse := DeployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}

		gock.New(baseURL).
			Post(triggerEndpoint).
			MatchHeader("Authorization", fmt.Sprintf("Bearer %s", apiToken)).
			MatchType("json").
			JSON(map[string]interface{}{
				"environment":             environment,
				"revision":                revision,
				"deployType":              DeployAll,
				"forceDeployWhenNoSemver": true,
			}).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
			DeployAll:   true,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		require.Empty(t, err)
		require.Equal(t, expectedResponse, deployResponse)

		require.True(t, gock.IsDone())
	})

	t.Run("failure", func(t *testing.T) {
		defer gock.Off()

		gock.New(baseURL).
			Post(triggerEndpoint).
			MatchHeader("Authorization", fmt.Sprintf("Bearer %s", apiToken)).
			Reply(400).
			JSON(map[string]interface{}{})

		cfg := DeployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := client.Trigger(projectId, cfg)
		base, _ := url.Parse(baseURL)
		path, _ := url.Parse(triggerEndpoint)
		require.EqualError(
			t,
			err,
			fmt.Sprintf("deploy error: POST %s: 400 - {}\n", base.ResolveReference(path)),
		)
		require.Empty(t, deployResponse)

		require.True(t, gock.IsDone())
	})
}
func TestStatusMonitor(t *testing.T) {
	const (
		projectId   = "u543t8sdf34t5"
		pipelineId  = 32562
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
		environment = "preprod"
	)

	client := testCreateDeployClientToken(t, baseURL, apiToken)

	t.Run("get status - success immediately", func(t *testing.T) {
		defer gock.Off()

		const expectedStatus = Success
		pipelinesTriggered := PipelinesConfig{
			PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
			PipelineConfig{
				ProjectId:   "dc24c12fe",
				PipelineId:  143295,
				Environment: environment,
			},
		}

		for _, p := range pipelinesTriggered {
			statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", p.ProjectId, p.PipelineId)
			gock.New(baseURL).
				Get(statusEndpoint).
				MatchParam("environment", p.Environment).
				Reply(200).
				JSON(map[string]interface{}{
					"id":     pipelineId,
					"status": expectedStatus,
				})
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := client.StatusMonitor(buf, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered), lastDeployedCompleted, "all the deploy were completed")
		require.Empty(t, slm.CallCount, "no need to wait")

		require.True(t, gock.IsDone())
	})

	t.Run("get status - pending -> running -> success", func(t *testing.T) {
		defer gock.Off()

		const finalStatus = Success
		const runningTimes = 2
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		pipelinesTriggered := PipelinesConfig{
			PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		pipelineStatuses := []PipelineStatus{Created, Pending, Running, Running, finalStatus}

		for _, ps := range pipelineStatuses {
			gock.New(baseURL).
				Get(statusEndpoint).
				MatchParam("environment", environment).
				Reply(200).
				JSON(map[string]interface{}{
					"id":     pipelineId,
					"status": ps,
				})
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := client.StatusMonitor(buf, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered), lastDeployedCompleted, "all the deploy were completed")
		require.Equal(t, len(pipelineStatuses)-1, slm.CallCount, "wait when created, pending and running received")

		require.True(t, gock.IsDone())
	})

	t.Run("get status - running -> failed", func(t *testing.T) {
		defer gock.Off()
		const finalStatus = Failed
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		pipelinesTriggered := PipelinesConfig{
			PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		pipelineStatuses := []PipelineStatus{Running, finalStatus}

		for _, ps := range pipelineStatuses {
			gock.New(baseURL).
				Get(statusEndpoint).
				MatchParam("environment", environment).
				Reply(200).
				JSON(map[string]interface{}{
					"id":     pipelineId,
					"status": ps,
				})
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := client.StatusMonitor(buf, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered), lastDeployedCompleted, "all the deploy were completed")
		require.Equal(t, 1, slm.CallCount, "wait once when running received")

		require.True(t, gock.IsDone())
	})

	t.Run("get status - error", func(t *testing.T) {
		defer gock.Off()
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": Created,
			})

		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(400).
			JSON(map[string]interface{}{})

		pipelinesTriggered := PipelinesConfig{
			PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := client.StatusMonitor(buf, &pipelinesTriggered, slm)

		require.Error(t, err)
		require.Contains(t, err.Error(), "status error:")
		require.Empty(t, lastDeployedCompleted, "no deploy was completed")
		require.Equal(t, 1, slm.CallCount, "wait only once")

		require.True(t, gock.IsDone())
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

func (sm *sleeperMock) Sleep() {
	sm.CallCount += 1
}
