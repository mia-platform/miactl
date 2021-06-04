package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	utils "github.com/mia-platform/miactl/cmd/internal"
	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCmd(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZOk2"
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
		environment = "test"
		pipelineId  = 457321
	)

	t.Run("get pipeline status with success - pipeline success", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		expectedStatuses := []sdk.PipelineStatus{
			sdk.Created,
			sdk.Pending,
			sdk.Running,
			sdk.Success,
		}

		callsCount := 0

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		for pid, status := range expectedStatuses {
			statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pid)

			handler := func(w http.ResponseWriter, r *http.Request) {
				callsCount += 1
				data, _ := json.Marshal(map[string]interface{}{
					"id":     pid,
					"status": status,
				})
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}
			server, _ := utils.CreateConfigurableTestServer(t, statusEndpoint, handler, nil)
			defer server.Close()

			viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
			viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

			cmd, buf, ctx := prepareStatusCmd(pid, "")
			require.NoError(t, cmd.ExecuteContext(ctx))

			tableRows := renderer.CleanTableRows(buf.String())

			expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
			expectedRow := fmt.Sprintf("%s | %d | %s", projectId, pid, status)
			require.Equal(t, expectedHeaders, tableRows[0])
			require.Equal(t, expectedRow, tableRows[1])
		}

		require.Equal(t, len(expectedStatuses), callsCount)
	})

	t.Run("get pipeline status with success - pipeline error", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := sdk.Failed
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{
				"id":     pipelineId,
				"status": expectedStatus,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, statusEndpoint, handler, nil)
		defer server.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineId, "")
		require.EqualError(t, cmd.ExecuteContext(ctx), "Deploy pipeline failed")

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, pipelineId, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		require.Equal(t, 1, callsCount)
	})

	t.Run("get pipeline status with success - set environment flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := sdk.Pending
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{
				"id":     pipelineId,
				"status": expectedStatus,
			})

			require.Equal(t, environment, r.FormValue("environment"))

			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, statusEndpoint, handler, nil)
		defer server.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineId, environment)
		require.NoError(t, cmd.ExecuteContext(ctx))

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, pipelineId, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		require.Equal(t, 1, callsCount)
	})

	t.Run("error getting pipeline status", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectId, pipelineId)
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, statusEndpoint, handler, nil)
		defer server.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineId, "")
		require.NoError(t, cmd.ExecuteContext(ctx))

		base, _ := url.Parse(server.URL)
		path, _ := url.Parse(statusEndpoint)
		require.Contains(
			t,
			buf.String(),
			fmt.Sprintf("GET %s: 400", base.ResolveReference(path)),
		)
		require.Equal(t, 1, callsCount)
	})

	t.Run("missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _, ctx := prepareStatusCmd(pipelineId, "")
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

		cmd, _, ctx := prepareStatusCmd(pipelineId, "")
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

		cmd, _, ctx := prepareStatusCmd(pipelineId, "")
		err := cmd.ExecuteContext(ctx)
		require.Contains(t, err.Error(), "no such flag -project")
	})
}

func prepareStatusCmd(pid int, environment string) (*cobra.Command, *bytes.Buffer, context.Context) {
	buf := &bytes.Buffer{}
	cmd := NewStatusCmd()

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{strconv.Itoa(pid)})
	if environment != "" {
		cmd.Flags().Set("environment", environment)
	}

	ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())

	return cmd, buf, ctx
}
