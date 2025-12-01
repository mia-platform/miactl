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
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mia-platform/miactl/internal/resources"
	"github.com/mia-platform/miactl/internal/util"
)

func rowForService(service resources.Service) []string {
	ports := make([]string, 0, len(service.Ports))
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
		strconv.Itoa(totalRestart),
		util.HumanDuration(time.Since(pod.Age)),
	}
}

func rowForJob(job resources.Job) []string {
	duration := "-"
	if !job.CompletionTime.IsZero() {
		duration = util.HumanDuration(job.CompletionTime.Sub(job.StartTime))
	}

	return []string{
		job.Name,
		fmt.Sprintf("%d/%d", job.Succeeded, (job.Active + job.Failed + job.Succeeded)),
		duration,
		util.HumanDuration(time.Since(job.Age)),
	}
}

func rowForDeployment(deployment resources.Deployment) []string {
	return []string{
		deployment.Name,
		fmt.Sprintf("%d/%d", deployment.Ready, deployment.Available),
		strconv.Itoa(deployment.Replicas),
		strconv.Itoa(deployment.Available),
		util.HumanDuration(time.Since(deployment.Age)),
	}
}

func rowForCronJob(cronjob resources.CronJob) []string {
	return []string{
		cronjob.Name,
		cronjob.Schedule,
		strconv.FormatBool(cronjob.Suspend),
		strconv.Itoa(cronjob.Active),
		util.HumanDuration(time.Since(cronjob.LastSchedule)),
		util.HumanDuration(time.Since(cronjob.Age)),
	}
}
