package deploy

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/internal/mocks"
	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk/deploy"

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
	expectedBearer := fmt.Sprintf("Bearer %s", apiToken)
	expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
	triggerEndpoint := fmt.Sprintf("/api/deploy/projects/%s/trigger/pipeline/", projectId)

	t.Run("successful deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: triggerEndpoint,
				Method:   http.MethodPost,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				RequestBody: deploy.DeployRequest{
					Environment:             environment,
					Revision:                revision,
					DeployType:              deploy.SmartDeploy,
					ForceDeployWhenNoSemver: false,
				},
				Reply: map[string]interface{}{
					"id":  expectedPipelineId,
					"url": expectedPipelineURL,
				},
				ReplyStatus: http.StatusOK,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, nil)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareCmd(t, environment, revision)
		err = cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | VIEW PIPELINE"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, expectedPipelineId, expectedPipelineURL)

		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])
	})

	t.Run("failed deploy", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: triggerEndpoint,
				Method:   http.MethodPost,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				RequestBody: deploy.DeployRequest{
					Environment:             environment,
					Revision:                revision,
					DeployType:              deploy.SmartDeploy,
					ForceDeployWhenNoSemver: false,
				},
				Reply:       map[string]interface{}{},
				ReplyStatus: http.StatusBadRequest,
			},
		}

		s, err := mocks.HTTPServer(t, mockConfigs, nil)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareCmd(t, environment, revision)
		err = cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		base, _ := url.Parse(s.URL)
		path, _ := url.Parse(triggerEndpoint)
		require.Contains(
			t,
			buf.String(),
			fmt.Sprintf("POST %s: 400", base.ResolveReference(path)),
		)
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
