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
		require.Equal(t, 0, len(pipelines))

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
