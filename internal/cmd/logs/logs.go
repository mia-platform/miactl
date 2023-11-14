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

package logs

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	logsEndpointTemplate = "/api/projects/%s/environments/%s/pods/logs"
	listEndpointTemplate = "/api/projects/%s/environments/%s/pods/describe/"
)

func Command(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs resource-query",
		Short: "Show logs related to a runtime resource using a regex query",
		Long:  "Show logs related to a runtime resource using a regex query.",

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			stream, err := getLogs(client, restConfig.ProjectID, restConfig.Environment, args[0], o.FollowLogs)
			cobra.CheckErr(err)

			defer stream.Close()
			_, err = io.Copy(cmd.OutOrStdout(), stream)
			cobra.CheckErr(err)
		},
	}

	flags := cmd.Flags()
	o.AddEnvironmentFlags(flags)
	o.AddLogsFlags(flags)

	return cmd
}

func getLogs(client *client.APIClient, projectID, environment, podRegex string, follow bool) (io.ReadCloser, error) {
	if projectID == "" {
		return nil, fmt.Errorf("missing project id, please set one with the flag or context")
	}

	if environment == "" {
		return nil, fmt.Errorf("missing environment, please set one with the flag or context")
	}

	regex, err := regexp.Compile(podRegex)
	if err != nil {
		return nil, err
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEndpointTemplate, projectID, environment)).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	pods := make([]resources.Pod, 0)
	if err := resp.ParseResponse(&pods); err != nil {
		return nil, err
	}

	logRequest := client.
		Get().
		APIPath(fmt.Sprintf(logsEndpointTemplate, projectID, environment)).
		SetHeader("Accept", "text/html").
		SetParam("file", "true").
		SetParam("follow", strconv.FormatBool(follow))

	for _, pod := range pods {
		if regex.MatchString(pod.Name) {
			containers := make([]string, len(pod.Containers))
			for _, container := range pod.Containers {
				containers = append(containers, container.Name)
			}
			logRequest.SetParam(pod.Name, containers...)
		}
	}

	return logRequest.Stream(context.Background())
}
