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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type deployRespnse struct {
	Url string
	Id  int
}

// Request is the body parameters needed to trigger a pipeline deploy.
type Request struct {
	Environment             string `json:"environment"`
	Revision                string `json:"revision"`
	DeployType              string `json:"deployType"`
	ForceDeployWhenNoSemver bool   `json:"forceDeployWhenNoSemver"`
}

type statusResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}
type initClient func(*clioptions.CLIOptions, string, string) (*httphandler.MiaClient, error)

var currentContext string

func NewDeployCmd(options *clioptions.CLIOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy project",
		Long:  "trigger the deploy pipeline for selected project",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.Get("current-context") != "" {
				currentContext = fmt.Sprint(viper.Get("current-context"))
				context.SetContextValues(cmd, currentContext)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			env := args[0]

			err := run(env, options, initializeClient)
			if err != nil {
				return err
			}
			return nil

		},
	}
	options.AddContextFlags(cmd)
	options.AddConnectionFlags(cmd)
	options.AddDeployFlags(cmd)
	return cmd

}

func run(env string, options *clioptions.CLIOptions, initializeClient initClient) error {
	epDeploy := fmt.Sprintf("/api/deploy/projects/%s/trigger/pipeline/", options.ProjectID)
	mcDeploy, err := initializeClient(options, epDeploy, currentContext)
	if err != nil {
		return fmt.Errorf("error generating the session: %w", err)
	}

	resp, err := triggerPipeline(mcDeploy, env, options)
	if err != nil {
		return fmt.Errorf("error executing the deploy request: %w", err)
	}
	fmt.Printf("Deploying project %s in the environment '%s'\n", options.ProjectID, env)

	epWait := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", options.ProjectID, resp.Id)
	mcWait, err := initializeClient(options, epWait, currentContext)
	if err != nil {
		return fmt.Errorf("error generating the session: %w", err)
	}

	status, err := waitStatus(mcWait)
	if err != nil {
		return fmt.Errorf("error retriving the pipeline status: %w", err)
	}
	fmt.Printf("Pipeline result: %s", status)
	return nil
}

func initializeClient(opts *clioptions.CLIOptions, endpoint string, currentContext string) (*httphandler.MiaClient, error) {
	client := httphandler.NewMiaClientBuilder()
	session, err := httphandler.ConfigureDefaultSessionHandler(opts, currentContext, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error building default session handler: %w", err)
	}

	client.WithSessionHandler(*session)

	return client, nil
}

func triggerPipeline(mc *httphandler.MiaClient, env string, options *clioptions.CLIOptions) (*deployRespnse, error) {
	data := Request{
		Environment:             env,
		Revision:                options.Revision,
		DeployType:              options.DeployType,
		ForceDeployWhenNoSemver: options.ForceDeployNoSemVer,
	}

	if options.DeployType == "deploy_all" {
		data.DeployType = options.DeployType
		data.ForceDeployWhenNoSemver = true
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error mashalling body: %w", err)
	}

	resp, err := mc.SessionHandler.Post(bytes.NewBuffer(dataJSON)).ExecuteRequest()
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pipeline exited with status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var body deployRespnse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil

}

func waitStatus(client *httphandler.MiaClient) (string, error) {
	status := statusResponse{}
	for {
		time.Sleep(2 * time.Second)
		resp, err := client.SessionHandler.Get().ExecuteRequest()
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("pipeline status not 200: %d", resp.StatusCode)
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&status)
		if err != nil {
			return "", err
		}
		if status.Status != "running" && status.Status != "pending" {
			break
		}

		fmt.Printf("The pipeline is %s..\n", status.Status)
	}
	return status.Status, nil
}
