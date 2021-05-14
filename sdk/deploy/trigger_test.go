package deploy

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

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
	triggerEndpoint := fmt.Sprintf("/deploy/projects/%s/trigger/pipeline/", projectId)

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

		deployResponse, err := Trigger(baseURL, apiToken, projectId, cfg)
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

		deployResponse, err := Trigger(baseURL, apiToken, projectId, cfg)
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

		deployResponse, err := Trigger(baseURL, apiToken, projectId, cfg)
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
