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

package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	listEndpointTemplate = "/api/projects/%s/environments/%s/services/describe/"
)

func Command(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage Mia-Platform Console project runtime service resources",
		Long: `Manage Mia-Platform Console project runtime service resources.

A project on Mia-Platform Console once deployed can have one or more service resources associcated with one or more
of its environments.
`,
	}

	// add sub commands
	cmd.AddCommand(
		listCmd(o),
	)

	return cmd
}

func listCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list ENVIRONMENT",
		Short: "List all services for a project in an environment",
		Long:  "List all services for a project in an environment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printServicesList(client, restConfig.ProjectID, args[0])
		},
	}

	return cmd
}

func printServicesList(client *client.APIClient, projectID, environment string) error {
	if projectID == "" {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}
	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEndpointTemplate, projectID, environment)).
		Do(context.Background())

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	services := make([]resources.Service, 0)
	err = resp.ParseResponse(&services)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		fmt.Printf("No services found for %s environment\n", environment)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Name", "Type", "Cluster-IP", "Port(s)", "Age"})

	if err != nil {
		return err
	}

	for _, service := range services {
		table.Append(rowForService(service))
	}

	table.Render()
	return nil
}

func rowForService(service resources.Service) []string {
	ports := make([]string, 0)
	for _, port := range service.Ports {
		ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
	}

	clusterIP := service.ClusterIP
	if len(clusterIP) == 0 {
		clusterIP = "<none>"
	}

	return []string{
		service.Name,
		service.Type,
		clusterIP,
		strings.Join(ports, ","),
		util.HumanDuration(time.Since(service.Age)),
	}
}
