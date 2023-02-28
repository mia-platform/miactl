// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploy

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/mocks"
	"github.com/mia-platform/miactl/old/renderer"
	"github.com/mia-platform/miactl/old/sdk/deploy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCmd(t *testing.T) {
	const (
		projectID      = "4h6UBlNiZOk2"
		baseURL        = "http://console-base-url/"
		apiToken       = "YWNjZXNzVG9rZW4="
		environment    = "test"
		pipelineID     = 457321
		serverCertPath = "../../../../testdata/server-cert.pem"
		serverKeyPath  = "../../../../testdata/server-key.pem"
		caCertPath     = "../../../../testdata/ca-cert.pem"
	)
	expectedBearer := fmt.Sprintf("Bearer %s", apiToken)

	t.Run("get pipeline status with success - pipeline success", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		expectedStatuses := []deploy.PipelineStatus{
			deploy.Created,
			deploy.Pending,
			deploy.Running,
			deploy.Success,
		}

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		for pid, status := range expectedStatuses {
			statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pid)

			mockConfigs := mocks.ServerConfigs{
				{
					Endpoint: statusEndpoint,
					Method:   http.MethodGet,
					RequestHeaders: map[string]string{
						"Authorization": expectedBearer,
					},
					Reply: map[string]interface{}{
						"id":     pid,
						"status": status,
					},
					ReplyStatus: http.StatusOK,
				},
			}

			s, err := mocks.HTTPServer(t, mockConfigs, nil)
			require.NoError(t, err, "mock must start correctly")
			defer s.Close()

			viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
			viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

			cmd, buf, ctx := prepareStatusCmd(pid, "")
			require.NoError(t, cmd.ExecuteContext(ctx))

			tableRows := renderer.CleanTableRows(buf.String())

			expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
			expectedRow := fmt.Sprintf("%s | %d | %s", projectID, pid, status)
			require.Equal(t, expectedHeaders, tableRows[0])
			require.Equal(t, expectedRow, tableRows[1])
		}
	})

	t.Run("get pipeline status with success - pipeline error", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := deploy.Failed
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				Reply: map[string]interface{}{
					"id":     pipelineID,
					"status": expectedStatus,
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
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineID, "")
		require.EqualError(t, cmd.ExecuteContext(ctx), "Deploy pipeline failed")

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectID, pipelineID, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])
	})

	t.Run("get pipeline status with success - insecure access", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := deploy.Running
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		serverCfg := mocks.CertificatesConfig{
			CertPath: serverCertPath,
			KeyPath:  serverKeyPath,
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				Reply: map[string]interface{}{
					"id":     pipelineID,
					"status": expectedStatus,
				},
				ReplyStatus: http.StatusOK,
			},
		}
		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineID, "")
		cmd.Flags().Set("insecure", "true")
		require.NoError(t, cmd.ExecuteContext(ctx))

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectID, pipelineID, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])
	})

	t.Run("get pipeline status with success - select custom CA certificate", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := deploy.Running
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		serverCfg := mocks.CertificatesConfig{
			CertPath: serverCertPath,
			KeyPath:  serverKeyPath,
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				Reply: map[string]interface{}{
					"id":     pipelineID,
					"status": expectedStatus,
				},
				ReplyStatus: http.StatusOK,
			},
		}
		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectID)
		viper.Set("ca-cert", caCertPath)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineID, "")
		require.NoError(t, cmd.ExecuteContext(ctx))

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectID, pipelineID, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])
	})

	t.Run("get pipeline status with success - set environment flag", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := deploy.Pending
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				QueryParams: map[string]interface{}{
					"environment": environment,
				},
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				Reply: map[string]interface{}{
					"id":     pipelineID,
					"status": expectedStatus,
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
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, buf, ctx := prepareStatusCmd(pipelineID, environment)
		require.NoError(t, cmd.ExecuteContext(ctx))

		tableRows := renderer.CleanTableRows(buf.String())

		expectedHeaders := "PROJECT ID | DEPLOY ID | STATUS"
		expectedRow := fmt.Sprintf("%s | %d | %s", projectID, pipelineID, expectedStatus)
		require.Equal(t, expectedHeaders, tableRows[0])
		require.Equal(t, expectedRow, tableRows[1])
	})

	t.Run("error getting pipeline status - certificate issue", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		expectedStatus := deploy.Running
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		serverCfg := mocks.CertificatesConfig{
			CertPath: serverCertPath,
			KeyPath:  serverKeyPath,
		}

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
				},
				Reply: map[string]interface{}{
					"id":     pipelineID,
					"status": expectedStatus,
				},
				ReplyStatus: http.StatusOK,
			},
		}
		s, err := mocks.HTTPServer(t, mockConfigs, &serverCfg)
		require.NoError(t, err, "mock must start correctly")
		defer s.Close()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apibaseurl", fmt.Sprintf("%s/", s.URL))
		viper.Set("apitoken", apiToken)
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _, ctx := prepareStatusCmd(pipelineID, "")
		err = cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Regexp(t, regexp.MustCompile("x509: certificate signed by unknown authority|certificate is not standards compliant"), err)
	})

	t.Run("error getting pipeline status", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)

		mockConfigs := mocks.ServerConfigs{
			{
				Endpoint: statusEndpoint,
				Method:   http.MethodGet,
				RequestHeaders: map[string]string{
					"Authorization": expectedBearer,
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
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _, ctx := prepareStatusCmd(pipelineID, "")
		err = cmd.ExecuteContext(ctx)
		require.Error(t, err)

		base, _ := url.Parse(s.URL)
		path, _ := url.Parse(statusEndpoint)
		require.Contains(
			t,
			err.Error(),
			fmt.Sprintf("GET %s: 400", base.ResolveReference(path)),
		)
	})

	t.Run("missing base url", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")

		viper.Set("apitoken", apiToken)
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _, ctx := prepareStatusCmd(pipelineID, "")
		err := cmd.ExecuteContext(ctx)
		require.EqualError(t, err, "API base URL not specified nor configured")
	})

	t.Run("missing api token", func(t *testing.T) {
		viper.Reset()
		defer viper.Reset()

		viper.SetConfigFile("/tmp/.miaplatformctl.yaml")
		viper.Set("apibaseurl", baseURL)
		viper.Set("project", projectID)
		viper.WriteConfigAs("/tmp/.miaplatformctl.yaml")

		cmd, _, ctx := prepareStatusCmd(pipelineID, "")
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

		cmd, _, ctx := prepareStatusCmd(pipelineID, "")
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
