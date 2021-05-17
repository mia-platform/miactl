package deploy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mia-platform/miactl/factory"
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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		require.NoError(t, cmd.ExecuteContext(ctx))

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

		const expectedStatus = sdk.Success
		pipelinesTriggered := sdk.PipelinesConfig{
			sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  pipelineId,
				Environment: environment,
			},
		}

		pipelineStatuses := []sdk.PipelineStatus{sdk.Pending, expectedStatus}

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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		require.NoError(t, cmd.ExecuteContext(ctx))

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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		require.NoError(t, cmd.ExecuteContext(ctx))

		require.Equal(t, "no deploy pipelines triggered found\n", buf.String())

		var pipelines sdk.PipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &pipelines))
		require.Equal(t, 0, len(pipelines))
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
