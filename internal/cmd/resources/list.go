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

package resources

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	CronJobResourceType  = "cronjob"
	CronJobsResourceType = "cronjobs"

	DeploymentResourceType  = "deployment"
	DeploymentsResourceType = "deployments"

	JobResourceType  = "job"
	JobsResourceType = "jobs"

	PodResourceType  = "pod"
	PodsResourceType = "pods"

	ServiceResourceType  = "service"
	ServicesResourceType = "services"

	listEndpointTemplate = "/api/projects/%s/environments/%s/%s/describe/"
)

var resourcesAvailable = []string{
	CronJobResourceType,
	CronJobsResourceType,
	DeploymentResourceType,
	DeploymentsResourceType,
	JobResourceType,
	JobsResourceType,
	PodResourceType,
	PodsResourceType,
	ServiceResourceType,
	ServicesResourceType,
}

var autocompletableResources = []string{
	CronJobsResourceType,
	DeploymentsResourceType,
	JobsResourceType,
	PodsResourceType,
	ServicesResourceType,
}

func APIResourcesCommand(_ *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-resources",
		Short: "List Mia-Platform Console supported runtime resources",
		Long:  "List Mia-Platform Console supported runtime resources.",

		Run: func(cmd *cobra.Command, args []string) {
			writer := cmd.OutOrStdout()
			fmt.Fprint(writer, "NAME")
			fmt.Fprintln(writer)
			for _, resource := range autocompletableResources {
				fmt.Fprint(writer, resource)
				fmt.Fprintln(writer)
			}
		},
	}

	return cmd
}

func ListCommand(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list RESOURCE-TYPE",
		Short: "List Mia-Platform Console runtime resources",
		Long: `List Mia-Platform Console runtime resources.

A project on Mia-Platform Console once deployed can have one or more resource
of different kinds associcated with one or more of its environments.

Use "miactl runtime api-resources" for a complete list of currently supported resources.`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return resourcesCompletions(args, toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printList(cmd.Context(), client, restConfig.ProjectID, args[0], restConfig.Environment)
		},
		Example: `# List all pods in current context
miactl runtime list pods

# List all service in 'development' environment
miactl runtime list services --environment development`,
	}

	o.AddEnvironmentFlags(cmd.Flags())

	return cmd
}

func resourcesCompletions(args []string, toComplete string) []string {
	resources := make([]string, 0)
	if len(args) > 0 {
		return resources
	}

	for _, resource := range autocompletableResources {
		if strings.HasPrefix(resource, toComplete) {
			resources = append(resources, resource)
		}
	}

	return resources
}

func printList(ctx context.Context, client *client.APIClient, projectID, resourceType, environment string) error {
	if projectID == "" {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	if environment == "" {
		return fmt.Errorf("missing environment, please set one with the flag or context")
	}

	if !slices.Contains(resourcesAvailable, resourceType) {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	if !strings.HasSuffix(resourceType, "s") {
		resourceType += "s"
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(listEndpointTemplate, projectID, environment, resourceType)).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	tableHeaders := make([]string, 0)
	canonicalType := ""
	var rows [][]string
	switch resourceType {
	case PodResourceType, PodsResourceType:
		tableHeaders = append(tableHeaders, "Status", "Name", "Application", "Ready", "Phase", "Restart", "Age")
		rows, err = rowsForResources[resources.Pod](resp, rowForPod)
		canonicalType = PodsResourceType
	case CronJobResourceType, CronJobsResourceType:
		tableHeaders = append(tableHeaders, "Name", "Schedule", "Suspend", "Active", "Last Schedule", "Age")
		rows, err = rowsForResources[resources.CronJob](resp, rowForCronJob)
		canonicalType = CronJobsResourceType
	case DeploymentResourceType, DeploymentsResourceType:
		tableHeaders = append(tableHeaders, "Name", "Ready", "Up-to-Date", "Available", "Age")
		rows, err = rowsForResources[resources.Deployment](resp, rowForDeployment)
		canonicalType = DeploymentsResourceType
	case JobResourceType, JobsResourceType:
		tableHeaders = append(tableHeaders, "Name", "Finished Pods", "Duration", "Age")
		rows, err = rowsForResources[resources.Job](resp, rowForJob)
		canonicalType = JobsResourceType
	case ServiceResourceType, ServicesResourceType:
		tableHeaders = append(tableHeaders, "Name", "Type", "Cluster-IP", "Port(s)", "Age")
		rows, err = rowsForResources[resources.Service](resp, rowForService)
		canonicalType = ServicesResourceType
	}
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		fmt.Printf("No %s found for %s environment\n", canonicalType, environment)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAutoWrapText(false)
	table.SetHeader(tableHeaders)
	table.AppendBulk(rows)
	table.Render()
	return nil
}
