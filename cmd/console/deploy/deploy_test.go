package deploy

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"
	"github.com/mia-platform/miactl/sdk/factory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestNewDeployCmd(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZO"
		revision    = "master"
		environment = "development"
		baseURL     = "http://console-base-url/"
		apiToken    = "YWNjZXNzVG9rZW4="
	)
	const expectedPipelineId = 458467
	expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
	triggerEndpoint := fmt.Sprintf("/deploy/projects/%s/trigger/pipeline/", projectId)

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
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", baseURL)
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectId)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		buf := &bytes.Buffer{}
		cmd := NewDeployCmd()

		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{environment, revision})
		cmd.InheritedFlags().Set("project", projectId)
		cmd.Flags().Set("environment", environment)
		cmd.Flags().Set("revision", revision)

		ctx := factory.WithValueTest(context.Background(), cmd.OutOrStdout(), clientMock)
		err := cmd.ExecuteContext(ctx)
		require.Nil(t, err)

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | VIEW PIPELINE"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectId, expectedPipelineId, expectedPipelineURL)

		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])

		lastDeployPipeline := viper.GetInt("project-deploy-pipeline")
		require.Equal(t, expectedPipelineId, lastDeployPipeline, "Pipeline id differs from expected")

		require.True(t, gock.IsDone())
	})

}

func TestDeploy(t *testing.T) {
	const (
		projectId   = "4h6UBlNiZO"
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

		expectedResponse := deployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}

		gock.New(baseURL).
			Post(triggerEndpoint).
			MatchHeader("Authorization", fmt.Sprintf("Bearer %s", apiToken)).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		cfg := deployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := deploy(baseURL, apiToken, projectId, &cfg)
		require.Empty(t, err)
		require.Equal(t, expectedResponse, deployResponse)

		require.True(t, gock.IsDone())
	})

	t.Run("success - with smart deploy", func(t *testing.T) {
		defer gock.Off()

		const expectedPipelineId = 458467
		expectedPipelineURL := fmt.Sprintf("https://pipeline-url/%d", expectedPipelineId)
		expectedResponse := deployResponse{
			Id:  expectedPipelineId,
			Url: expectedPipelineURL,
		}

		gock.New(baseURL).
			Post(triggerEndpoint).
			MatchHeader("Authorization", fmt.Sprintf("Bearer %s", apiToken)).
			Reply(200).
			JSON(map[string]interface{}{
				"id":  expectedPipelineId,
				"url": expectedPipelineURL,
			})

		cfg := deployConfig{
			Environment:         environment,
			Revision:            revision,
			SmartDeploy:         true,
			ForceDeployNoSemVer: false,
		}

		deployResponse, err := deploy(baseURL, apiToken, projectId, &cfg)
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

		cfg := deployConfig{
			Environment: environment,
			Revision:    revision,
		}

		deployResponse, err := deploy(baseURL, apiToken, projectId, &cfg)
		require.EqualError(
			t,
			err,
			fmt.Sprintf("deploy error: POST http://console-base-url/deploy/projects/%s/trigger/pipeline/: 400 - {}\n", projectId),
		)
		require.Empty(t, deployResponse)

		require.True(t, gock.IsDone())
	})
}

func clientMock(opts sdk.Options) (*sdk.MiaClient, error) {
	return nil, nil
}
