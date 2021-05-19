package deploy

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewStatusCmd(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZOk2"
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
		environment = "test"
		pipelineId  = 457321
	)

	t.Run("get pipeline status with success - test all values", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		expectedStatuses := []sdk.PipelineStatus{
			sdk.Created,
			sdk.Pending,
			sdk.Running,
			sdk.Success,
			sdk.Failed,
			sdk.Canceled,
		}

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		for pid, status := range expectedStatuses {
			statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pid)
			gock.New(baseURL).
				Get(statusEndpoint).
				Reply(200).
				JSON(map[string]interface{}{
					"id":     pid,
					"status": status,
				})

			cmd, buf := prepareStatusCmd(pid, "")

			ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
			require.NoError(t, cmd.ExecuteContext(ctx))

			tableRows := renderer.CleanTableRows(buf.String())

			expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
			expectedRow := fmt.Sprintf("%s | %d | %s", projectId, pid, status)
			require.Equal(t, expectedHeaders, tableRows[0])
			require.Equal(t, expectedRow, tableRows[1])
		}

		require.True(t, gock.IsDone())
	})

	t.Run("get pipeline status with success - set environment flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		expectedStatus := sdk.Pending

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)
		gock.New(baseURL).
			Get(statusEndpoint).
			MatchParam("environment", environment).
			Reply(200).
			JSON(map[string]interface{}{
				"id":     pipelineId,
				"status": expectedStatus,
			})

		cmd, buf := prepareStatusCmd(pipelineId, environment)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		require.NoError(t, cmd.ExecuteContext(ctx))

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, pipelineId, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		require.True(t, gock.IsDone())
	})

	t.Run("error getting pipeline status", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		defer gock.Off()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)
		gock.New(baseURL).
			Get(statusEndpoint).
			Reply(400).
			JSON(map[string]interface{}{})

		cmd, buf := prepareStatusCmd(pipelineId, "")

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		require.NoError(t, cmd.ExecuteContext(ctx))

		base, _ := url.Parse(baseURL)
		path, _ := url.Parse(statusEndpoint)
		require.Contains(
			t,
			buf.String(),
			fmt.Sprintf("GET %s: 400", base.ResolveReference(path)),
		)

		require.True(t, gock.IsDone())
	})

	t.Run("missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _ := prepareStatusCmd(pipelineId, environment)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")
	})

	t.Run("missing api token", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _ := prepareStatusCmd(pipelineId, environment)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "missing API token - please login")
	})

	t.Run("missing project flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _ := prepareStatusCmd(pipelineId, environment)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.Contains(t, err.Error(), "no such flag -project")
	})
}

func prepareStatusCmd(pid int, environment string) (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := NewStatusCmd()

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{strconv.Itoa(pid)})
	if environment != "" {
		cmd.Flags().Set("environment", environment)
	}

	return cmd, buf
}
