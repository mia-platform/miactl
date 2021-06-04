package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	utils "github.com/mia-platform/miactl/cmd/internal"
	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/renderer"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
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
	triggerEndpoint := fmt.Sprintf("/api/deploy/projects/%s/trigger/pipeline/", projectId)

	t.Run("successful deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, triggerEndpoint, handler, nil)
		defer server.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareCmd(t, environment, revision)
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | VIEW PIPELINE"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, expectedPipelineId, expectedPipelineURL)

		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		require.Equal(t, 1, callsCount)
	})

	t.Run("failed deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		callsCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callsCount += 1
			data, _ := json.Marshal(map[string]interface{}{})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
		}
		server, _ := utils.CreateConfigurableTestServer(t, triggerEndpoint, handler, nil)
		defer server.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", server.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareCmd(t, environment, revision)
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		base, _ := url.Parse(server.URL)
		path, _ := url.Parse(triggerEndpoint)
		require.Contains(
			t,
			buf.String(),
			fmt.Sprintf("POST %s: 400", base.ResolveReference(path)),
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

		cmd, _, ctx := prepareCmd(t, environment, revision)
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

		cmd, _, ctx := prepareCmd(t, environment, revision)
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

		cmd, _, ctx := prepareCmd(t, environment, revision)
		err := cmd.ExecuteContext(ctx)
		require.Contains(t, err.Error(), "no such flag -project")
	})

	t.Run("missing environment flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.Regexp(t, ".*environment.* not set", err.Error())
	})

	t.Run("missing revision flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

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

		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		err := cmd.ExecuteContext(ctx)
		require.Regexp(t, ".*revision.* not set", err.Error())
	})
}

func prepareCmd(t *testing.T, environment, revision string) (*cobra.Command, *bytes.Buffer, context.Context) {
	t.Helper()

	buf := &bytes.Buffer{}
	cmd := NewDeployCmd()

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Flags().Set("environment", environment)
	cmd.Flags().Set("revision", revision)

	ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())

	return cmd, buf, ctx
}
