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
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	logsEndpointTemplate = "/api/projects/%s/environments/%s/pods/logs"
	listEndpointTemplate = "/api/projects/%s/environments/%s/pods/describe/"
)

func Command(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs resource-query",
		Short: "Show logs related to a runtime resource using a regex query",
		Long: `Show logs related to a runtime resource using a regex query.

You can write any regex compatible with RE2 excluding -C. The regex than will
be used to filter down the list of pods available in the current context and
then the logs of all their containers will be displayed.`,

		Example: `# Get all logs for pods that begin with api-gateway
miactl runtime logs api-gateway

# Get all logs for pods named exactly job-name
miactl runtime logs "^job-name$"`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			stream, err := getLogs(cmd.Context(), client, restConfig.ProjectID, restConfig.Environment, args[0], o.FollowLogs)
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

func getLogs(ctx context.Context, client *client.APIClient, projectID, environment, podRegex string, follow bool) (io.ReadCloser, error) {
	if projectID == "" {
		return nil, errors.New("missing project id, please set one with the flag or context")
	}

	if environment == "" {
		return nil, errors.New("missing environment, please set one with the flag or context")
	}

	regex, err := regexp.Compile(podRegex)
	if err != nil {
		return nil, err
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEndpointTemplate, projectID, environment)).
		Do(ctx)

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
			containers := make([]string, 0, len(pod.Containers))
			for _, container := range pod.Containers {
				containers = append(containers, container.Name)
			}
			logRequest.SetParam(pod.Name, containers...)
		}
	}

	return logRequest.Stream(ctx)
}
