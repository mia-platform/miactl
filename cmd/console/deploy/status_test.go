package deploy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewStatusCmd(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZOk2"
		pipelineId  = 564745
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
		environment = "test"
	)
	statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

	t.Run("get status with success", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		const expectedStatus = "success"
		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": expectedStatus,
			})

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.Set(triggeredPipelinesKey, pipelinesTriggered)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewStatusCmd()
		// needed since the test command does not inherit root command settings
		cmd.SilenceUsage = true

		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.ExecuteContext(context.Background())
		require.NoError(t, err)

		outputLines := strings.Split(buf.String(), "\n")

		// account for one more line that inform that all deploy were completed
		require.Equal(t, len(pipelinesTriggered), len(sliceFilter(t, outputLines))-1)
		for idx, pipeline := range pipelinesTriggered {
			require.Equal(
				t,
				fmt.Sprintf("project: %s\tpipeline: %d\tstatus:%s", pipeline.ProjectId, pipeline.PipelineId, expectedStatus),
				outputLines[idx],
			)
		}

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Empty(t, pipelines)

		require.True(t, gock.IsDone())
	})

	t.Run("get status - failed to obtain ", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(400).
			JSON(map[string]interface{}{})

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.Set(triggeredPipelinesKey, pipelinesTriggered)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewStatusCmd()
		// needed since the test command does not inherit root command settings
		cmd.SilenceUsage = true

		cmd.SetOut(buf)
		cmd.SetErr(buf)

		err := cmd.ExecuteContext(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "status error:")

		outputLines := strings.Split(buf.String(), "\n")

		require.Equal(t, 1, len(sliceFilter(t, outputLines)))

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Equal(t, 1, len(pipelines))
		require.Equal(t, pipelinesTriggered, pipelines)

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
	statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

	t.Run("get status - success immediately", func(t *testing.T) {
		const expectedStatus = "success"
		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": expectedStatus,
			})

		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := statusMonitor(buf, baseURL, apiToken, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered)-1, lastDeployedCompleted, "all the deploy were completed")
		require.Empty(t, slm.CallCount, "no need to wait")

		require.True(t, gock.IsDone())
	})

	t.Run("get status - pending -> running -> success", func(t *testing.T) {
		const finalStatus = Success
		const runningTimes = 2

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
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": Pending,
			})

		gock.New(baseURL).
			Get(statusEndpoint).
			Times(runningTimes).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": Running,
			})

		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": finalStatus,
			})

		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		buf := &bytes.Buffer{}
		slm := &sleeperMock{}

		lastDeployedCompleted, err := statusMonitor(buf, baseURL, apiToken, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered)-1, lastDeployedCompleted, "all the deploy were completed")
		require.Equal(t, 4, slm.CallCount, "wait when created, pending and running received")

		require.True(t, gock.IsDone())
	})
}

func sliceFilter(t testing.TB, s []string) []string {
	t.Helper()
	filtered := []string{}

	for _, e := range s {
		if e != "" {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

type sleeperMock struct {
	CallCount int
}

func (sm *sleeperMock) Sleep() {
	sm.CallCount += 1
}
