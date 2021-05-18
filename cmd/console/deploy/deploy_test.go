package deploy

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/renderer"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewDeployCmd(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZOk2"
		revision    = "master"
		environment = "development"
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
	)
	const expectedPipelineId = 458467
	expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
	triggerEndpoint := fmt.Sprintf("api/deploy/projects/%s/trigger/pipeline/", projectId)

	expectedPipeline := pipelineConfig{
		ProjectId:   projectId,
		PipelineId:  expectedPipelineId,
		Environment: environment,
	}

	t.Run("successful deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		gock.New(baseURL).
			Post(triggerEndpoint).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | VIEW PIPELINE"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, expectedPipelineId, expectedPipelineURL)

		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		var triggeredPipelines pipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &triggeredPipelines))
		require.Equal(t, 1, len(triggeredPipelines), "Number of triggered pipelines should match")
		require.Equal(t, expectedPipeline, triggeredPipelines[0], "pipelineConfig details should match")

		require.True(t, gock.IsDone())
	})

	t.Run("successful deploy - add new pipeline to list of triggered ones", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		gock.New(baseURL).
			Post(triggerEndpoint).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		// define from where login command should read config
		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.Set(triggeredPipelinesKey, pipelinesConfig{
			pipelineConfig{ProjectId: "437t34b293u", PipelineId: 723531, Environment: "test"},
		})
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | VIEW PIPELINE"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, expectedPipelineId, expectedPipelineURL)

		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		var triggeredPipelines pipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &triggeredPipelines))
		require.Equal(t, 2, len(triggeredPipelines), "Number of triggered pipelines should match")
		require.Equal(t, expectedPipeline, triggeredPipelines[1], "Last pipeline details should match")

		require.True(t, gock.IsDone())
	})

	t.Run("failed deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		gock.New(baseURL).
			Post(triggerEndpoint).
			Reply(400).
			JSON(map[string]interface{}{})

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		base, _ := url.Parse(baseURL)
		path, _ := url.Parse(triggerEndpoint)
		require.Contains(
			t,
			buf.String(),
			fmt.Sprintf("POST %s: 400", base.ResolveReference(path)),
		)

		var triggeredPipelines pipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &triggeredPipelines))
		require.Empty(t, triggeredPipelines, "No pipelines should be triggered")

		require.True(t, gock.IsDone())
	})

	t.Run("missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")

		var triggeredPipelines pipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &triggeredPipelines))
		require.Empty(t, triggeredPipelines, "No pipelines should be triggered")
	})

	t.Run("missing api token", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "missing API token - please login")

		var triggeredPipelines pipelinesConfig
		require.NoError(t, viper.UnmarshalKey(triggeredPipelinesKey, &triggeredPipelines))
		require.Empty(t, triggeredPipelines, "No pipelines should be triggered")
	})
}
