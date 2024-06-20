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

package events

import (
	"context"
	"fmt"
	"time"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/printer"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/util"
	"github.com/spf13/cobra"
)

const (
	eventsEndpointTemplate = "/api/projects/%s/environments/%s/resources/%s/events"
)

func Command(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events RESOURCE-NAME",
		Short: "Show events related to a runtime resource in a Mia-Platform Console project environment",
		Long:  "Show events related to a runtime resource in a Mia-Platform Console project environment.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return printEventsList(cmd.Context(), client, restConfig.ProjectID, restConfig.Environment, args[1], o.Printer())
		},
	}
	return cmd
}

func printEventsList(ctx context.Context, client *client.APIClient, projectID, environment, resourceName string, p printer.IPrinter) error {
	if projectID == "" {
		return fmt.Errorf("missing project id, please set one with the flag or context")
	}

	if environment == "" {
		return fmt.Errorf("missing environment, please set one with the flag or context")
	}

	resp, err := client.
		Get().
		APIPath(fmt.Sprintf(eventsEndpointTemplate, projectID, environment, resourceName)).
		Do(ctx)

	if err != nil {
		return err
	}

	if err := resp.Error(); err != nil {
		return err
	}

	events := make([]resources.RuntimeEvent, 0)
	err = resp.ParseResponse(&events)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Printf("No events found for %s in %s environment\n", resourceName, environment)
		return nil
	}

	p.Keys("Last Seen", "Type", "Reason", "Object", "Message")
	for _, event := range events {
		p.Record(rowForEvent(event)...)
	}

	p.Print()
	return nil
}

func rowForEvent(event resources.RuntimeEvent) []string {
	age := "-"
	if !event.LastSeen.IsZero() {
		age = util.HumanDuration(time.Since(event.LastSeen))
	} else if !event.FirstSeen.IsZero() {
		age = util.HumanDuration(time.Since(event.FirstSeen))
	}
	return []string{
		age,
		event.Type,
		event.Reason,
		event.Object,
		event.Message,
	}
}
