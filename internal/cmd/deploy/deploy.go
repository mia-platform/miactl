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
)

type deployResponse struct {
	URL string
	ID  int
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
		Use:   "deploy ENVIRONMENT",
		Short: "deploy project",
		Long:  "trigger the deploy pipeline for selected project",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			currentContext, err := context.GetCurrentContext()
			if err != nil {
				return err
			}
			return context.SetContextValues(cmd, currentContext)
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
	options.AddConnectionFlags(cmd.PersistentFlags())
	options.AddContextFlags(cmd.PersistentFlags())
	options.AddCompanyFlags(cmd.PersistentFlags())
	options.AddProjectFlags(cmd.PersistentFlags())
	options.AddDeployFlags(cmd.PersistentFlags())
	if err := cmd.MarkPersistentFlagRequired("revision"); err != nil {
		// if there is an error something very wrong is happening, panic
		panic(err)
	}
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

	epWait := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", options.ProjectID, resp.ID)
	mcWait, err := initializeClient(options, epWait, currentContext)
	if err != nil {
		return fmt.Errorf("error generating the session: %w", err)
	}

	status, err := waitStatus(mcWait)
	if err != nil {
		return fmt.Errorf("error retrieving the pipeline status: %w", err)
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

func triggerPipeline(mc *httphandler.MiaClient, env string, options *clioptions.CLIOptions) (*deployResponse, error) {
	data := Request{
		Environment:             env,
		Revision:                options.Revision,
		DeployType:              options.DeployType,
		ForceDeployWhenNoSemver: options.NoSemVer,
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

	var body deployResponse

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
