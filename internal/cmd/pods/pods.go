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

package pods

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	listEndpointTemplate = "/api/projects/%s/environments/%s/pods/describe/"
)

func PodCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pod",
		Short: "Manage Mia-Platform Console project runtime pod resources",
		Long: `Manage Mia-Platform Console project runtime pod resources.

A project on Mia-Platform Console once deployed can have one or more pod resources associcated with one or more
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
		Use:   "list [environment]",
		Short: "List all pods for a project in an environment",
		Long:  "List all pods for a project in an environment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printPodsList(client, restConfig.ProjectID, args[0])
		},
	}

	return cmd
}

func printPodsList(client *client.APIClient, projectID, environment string) error {
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

	pods := make([]resources.Pod, 0)
	err = resp.ParseResponse(&pods)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		fmt.Printf("No pods found for %s environment", environment)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeader([]string{"Status", "Name", "Application", "Ready", "Phase", "Restart", "Age"})

	if err != nil {
		return err
	}

	for _, pod := range pods {
		table.Append(rowForPod(pod))
	}

	table.Render()
	return nil
}

func rowForPod(pod resources.Pod) []string {
	totalRestart := 0
	totalContainers := 0
	readyContainers := 0
	for _, container := range pod.Containers {
		totalRestart += container.RestartCount
		totalContainers++
		if container.Ready {
			readyContainers++
		}
	}

	components := make([]string, 0)
	for _, component := range pod.Component {
		if len(component.Name) == 0 {
			continue
		}

		nameComponents := []string{component.Name}
		if len(component.Version) > 0 {
			nameComponents = append(nameComponents, component.Version)
		}
		components = append(components, strings.Join(nameComponents, ":"))
	}

	if len(components) == 0 {
		components = append(components, "-")
	}

	caser := cases.Title(language.English)
	return []string{
		caser.String(pod.Status),
		pod.Name,
		strings.Join(components, ", "),
		fmt.Sprintf("%d/%d", readyContainers, totalContainers),
		caser.String(pod.Phase),
		fmt.Sprint(totalRestart),
		HumanDuration(time.Since(pod.StartTime)),
	}
}

// adapted from https://github.com/kubernetes/apimachinery/blob/master/pkg/util/duration/duration.go#L48
// HumanDuration returns a succinct representation of the provided duration with limited precision for
// consumption by humans. It provides ~2-3 significant figures of duration.
func HumanDuration(d time.Duration) string {
	var durationString string

	switch {
	case d < time.Minute*2: // show seconds duration until 2 minutes
		convertedDuration := d.Round(time.Second) / time.Second
		durationString = fmt.Sprintf("%ds", convertedDuration)
	case d < time.Minute*10: // show minutes and seconds duration until 10 minutes
		convertedDuration := d.Round(time.Second)
		durationString = fmt.Sprint(convertedDuration)
	case d < time.Hour*3: // show minutes duration until 3 hours
		convertedDuration := d.Round(time.Second) / time.Minute
		durationString = fmt.Sprintf("%dm", convertedDuration)
	case d < time.Hour*8: // show hours and minutes duration until 8 hours
		convertedDuration := d.Round(time.Minute) / time.Minute
		durationString = fmt.Sprintf("%dh%dm", convertedDuration/60, convertedDuration%60)
	case d < time.Hour*48: // show hours duration until 2 days
		convertedDuration := d.Round(time.Minute) / time.Hour
		durationString = fmt.Sprintf("%dh", convertedDuration)
	case d < time.Hour*192: // show days and hours duration until ~ 8 days (24 h * 8 days = 192 hours)
		convertedDuration := d.Round(time.Minute) / time.Hour
		residualHours := convertedDuration % 24
		if residualHours == 0 {
			durationString = fmt.Sprintf("%dd", convertedDuration/24)
		} else {
			durationString = fmt.Sprintf("%dd%dh", convertedDuration/24, residualHours)
		}
	case d < time.Hour*8760: // show days duration until ~ 1 year (24 h * 365 days = 8760 hours)
		convertedDuration := d.Round(time.Hour) / time.Hour
		durationString = fmt.Sprintf("%dd", convertedDuration/24)
	default: // show days and years duration after the first year
		convertedDuration := (d.Round(time.Hour) / time.Hour) / 24
		residualDays := convertedDuration % 365
		if residualDays == 0 {
			durationString = fmt.Sprintf("%dy", convertedDuration/365)
		} else {
			durationString = fmt.Sprintf("%dy%dd", convertedDuration/365, residualDays)
		}
	}

	return durationString
}
