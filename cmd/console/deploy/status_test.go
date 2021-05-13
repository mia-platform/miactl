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

		rawOutputLines := strings.Split(buf.String(), "\n")
		outputLines := sliceFilter(t, rawOutputLines)

		// account for one more line that inform that all deploy were completed
		require.Equal(t, len(pipelinesTriggered), len(outputLines)-1)
		for idx, pipeline := range pipelinesTriggered {
			require.Equal(
				t,
				fmt.Sprintf("project: %s\tpipeline: %d\tstatus:%s", pipeline.ProjectId, pipeline.PipelineId, expectedStatus),
				outputLines[idx],
			)
		}
		require.Equal(t, endMessage, outputLines[len(sliceFilter(t, outputLines))-1])

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Empty(t, pipelines)

		require.True(t, gock.IsDone())
	})

	t.Run("get status with success after pending", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		const expectedStatus = Success
		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		pipelineStatuses := []PipelineStatus{Pending, expectedStatus}

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

		rawOutputLines := strings.Split(buf.String(), "\n")
		outputLines := sliceFilter(t, rawOutputLines)

		// account for one more line that inform that all deploy were completed
		require.Equal(t, len(pipelinesTriggered), len(outputLines)-1)
		for idx, pipeline := range pipelinesTriggered {
			require.Equal(
				t,
				fmt.Sprintf("project: %s\tpipeline: %d\tstatus:%s", pipeline.ProjectId, pipeline.PipelineId, expectedStatus),
				outputLines[idx],
			)
		}
		require.Equal(t, endMessage, outputLines[len(sliceFilter(t, outputLines))-1])

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Empty(t, pipelines)

		require.True(t, gock.IsDone())
	})

	t.Run("get status - failed to obtain it", func(t *testing.T) {
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

	t.Run("get status - missing API token", func(t *testing.T) {
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

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
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
		require.EqualError(t, err, "missing API token - please login")

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Equal(t, 1, len(pipelines))
		require.Equal(t, pipelinesTriggered, pipelines)
	})

	t.Run("get status - missing base url", func(t *testing.T) {
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

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

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
		require.EqualError(t, err, "API base URL not specified nor configured")

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Equal(t, 1, len(pipelines))
		require.Equal(t, pipelinesTriggered, pipelines)
	})

	t.Run("get status - no pipeline previously triggered", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewStatusCmd()
		// needed since the test command does not inherit root command settings
		cmd.SilenceUsage = true

		cmd.SetOut(buf)
		cmd.SetErr(buf)

		require.NoError(t, cmd.ExecuteContext(context.Background()))

		require.Equal(t, "no deploy pipelines triggered found\n", buf.String())

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Equal(t, 0, len(pipelines))
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

	t.Run("get status - success immediately", func(t *testing.T) {
		defer gock.Off()

		const expectedStatus = Success
		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
			sdk.PipelineConfig{
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

		lastDeployedCompleted, err := statusMonitor(buf, baseURL, apiToken, &pipelinesTriggered, slm)

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

		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
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

		lastDeployedCompleted, err := statusMonitor(buf, baseURL, apiToken, &pipelinesTriggered, slm)

		require.NoError(t, err)
		require.Equal(t, len(pipelinesTriggered), lastDeployedCompleted, "all the deploy were completed")
		require.Equal(t, len(pipelineStatuses)-1, slm.CallCount, "wait when created, pending and running received")

		require.True(t, gock.IsDone())
	})

	t.Run("get status - running -> failed", func(t *testing.T) {
		defer gock.Off()
		const finalStatus = Failed
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)

		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
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

		lastDeployedCompleted, err := statusMonitor(buf, baseURL, apiToken, &pipelinesTriggered, slm)

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

		require.Error(t, err)
		require.Contains(t, err.Error(), "status error:")
		require.Empty(t, lastDeployedCompleted, "no deploy was completed")
		require.Equal(t, 1, slm.CallCount, "wait only once")

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
