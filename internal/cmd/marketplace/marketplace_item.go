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

package marketplace

import "github.com/mia-platform/miactl/internal/encoding"

func Unmarshal(data []byte, encodingFormat string) (Item, error) {
	var r Item

	err := encoding.UnmarshalData(data, encodingFormat, &r)
	if err != nil {
		return Item{}, err
	}
	return r, nil
}

func (r *Item) Marshal(encodingFormat string) ([]byte, error) {
	return encoding.MarshalData(r, encodingFormat, encoding.MarshalOptions{Indent: true})
}

type Item struct {
	ID                        *string                  `json:"_id,omitempty"  yaml:"_id,omitempty"` //nolint:tagliatelle
	Category                  *Category                `json:"category,omitempty"  yaml:"category,omitempty"`
	CategoryID                *string                  `json:"categoryId,omitempty"  yaml:"categoryId,omitempty"`
	ComingSoon                *bool                    `json:"comingSoon,omitempty"  yaml:"comingSoon,omitempty"`
	ComponentsIDS             []string                 `json:"componentsIds,omitempty"  yaml:"componentsIds,omitempty"`
	Description               *string                  `json:"description,omitempty"  yaml:"description,omitempty"`
	Documentation             []map[string]interface{} `json:"documentation,omitempty"  yaml:"documentation,omitempty"`
	ImageURL                  *string                  `json:"imageUrl,omitempty"  yaml:"imageUrl,omitempty"`
	Name                      *string                  `json:"name,omitempty"  yaml:"name,omitempty"`
	PublishOnMiaDocumentation *bool                    `json:"publishOnMiaDocumentation,omitempty"  yaml:"publishOnMiaDocumentation,omitempty"`
	ProviderID                *string                  `json:"providerId,omitempty"  yaml:"providerId,omitempty"`
	ReleaseStage              *string                  `json:"releaseStage,omitempty"  yaml:"releaseStage,omitempty"`
	RepositoryURL             *string                  `json:"repositoryUrl,omitempty"  yaml:"repositoryUrl,omitempty"`
	Resources                 *Resources               `json:"resources,omitempty"  yaml:"resources,omitempty"`
	SupportedBy               *string                  `json:"supportedBy,omitempty"  yaml:"supportedBy,omitempty"`
	SupportedByImageURL       *string                  `json:"supportedByImageUrl,omitempty"  yaml:"supportedByImageUrl,omitempty"`
	Type                      *string                  `json:"type,omitempty"  yaml:"type,omitempty"`
}

type Category struct {
	ID    string `json:"id"  yaml:"id"`
	Label string `json:"label"  yaml:"label"`
}

type Resources struct {
	Collections         map[string]interface{} `json:"collections,omitempty"  yaml:"collections,omitempty"` // TODO: define this interface if needed
	Endpoints           map[string]interface{} `json:"endpoints,omitempty"  yaml:"endpoints,omitempty"`     // TODO: define this interface if needed
	Services            map[string]Service     `json:"services,omitempty"  yaml:"services,omitempty"`
	UnsecretedVariables map[string]interface{} `json:"unsecretedVariables,omitempty"  yaml:"unsecretedVariables,omitempty"` // TODO: define this interface if needed
}

type Service struct {
	AdditionalContainers                 []AdditionalContainer      `json:"additionalContainers,omitempty"  yaml:"additionalContainers,omitempty"`
	ArchiveURL                           *string                    `json:"archiveUrl,omitempty"  yaml:"archiveUrl,omitempty"`
	ComponentID                          *string                    `json:"componentId,omitempty"  yaml:"componentId,omitempty"`
	ContainerPorts                       []DefaultContainerPort     `json:"containerPorts,omitempty"  yaml:"containerPorts,omitempty"`
	CustomFilesConfig                    []DefaultCustomFilesConfig `json:"customFilesConfig,omitempty"  yaml:"customFilesConfig,omitempty"`
	DefaultAnnotations                   []DefaultAnnotation        `json:"defaultAnnotations,omitempty"  yaml:"defaultAnnotations,omitempty"`
	DefaultArgs                          []string                   `json:"defaultArgs,omitempty"  yaml:"defaultArgs,omitempty"`
	DefaultConfigMaps                    []DefaultConfigMap         `json:"defaultConfigMaps,omitempty"  yaml:"defaultConfigMaps,omitempty"`
	DefaultDocumentationPath             *string                    `json:"defaultDocumentationPath,omitempty"  yaml:"defaultDocumentationPath,omitempty"`
	DefaultEnvironmentVariables          []map[string]interface{}   `json:"defaultEnvironmentVariables,omitempty"  yaml:"defaultEnvironmentVariables,omitempty"`
	DefaultHeaders                       []DefaultHeader            `json:"defaultHeaders,omitempty"  yaml:"defaultHeaders,omitempty"`
	DefaultLabels                        []DefaultLabel             `json:"defaultLabels,omitempty"  yaml:"defaultLabels,omitempty"`
	DefaultLogParser                     *string                    `json:"defaultLogParser,omitempty"  yaml:"defaultLogParser,omitempty"`
	DefaultMonitoring                    *DefaultMonitoring         `json:"defaultMonitoring,omitempty"  yaml:"defaultMonitoring,omitempty"`
	DefaultProbes                        *DefaultProbes             `json:"defaultProbes,omitempty"  yaml:"defaultProbes,omitempty"`
	DefaultResources                     *DefaultResources          `json:"defaultResources,omitempty"  yaml:"defaultResources,omitempty"`
	DefaultSecrets                       []DefaultSecret            `json:"defaultSecrets,omitempty"  yaml:"defaultSecrets,omitempty"`
	DefaultTerminationGracePeriodSeconds *float64                   `json:"defaultTerminationGracePeriodSeconds,omitempty"  yaml:"defaultTerminationGracePeriodSeconds,omitempty"`
	Name                                 *string                    `json:"name,omitempty"  yaml:"name,omitempty"`
	Type                                 *string                    `json:"type,omitempty"  yaml:"type,omitempty"`
}

type AdditionalContainer struct {
	ArchiveURL                           *string                    `json:"archiveUrl,omitempty"  yaml:"archiveUrl,omitempty"`
	ComponentID                          *string                    `json:"componentId,omitempty"  yaml:"componentId,omitempty"`
	ContainerPorts                       []DefaultContainerPort     `json:"containerPorts,omitempty"  yaml:"containerPorts,omitempty"`
	CustomFilesConfig                    []DefaultCustomFilesConfig `json:"customFilesConfig,omitempty"  yaml:"customFilesConfig,omitempty"`
	DefaultAnnotations                   []DefaultAnnotation        `json:"defaultAnnotations,omitempty"  yaml:"defaultAnnotations,omitempty"`
	DefaultArgs                          []string                   `json:"defaultArgs,omitempty"  yaml:"defaultArgs,omitempty"`
	DefaultConfigMaps                    []DefaultConfigMap         `json:"defaultConfigMaps,omitempty"  yaml:"defaultConfigMaps,omitempty"`
	DefaultDocumentationPath             *string                    `json:"defaultDocumentationPath,omitempty"  yaml:"defaultDocumentationPath,omitempty"`
	DefaultEnvironmentVariables          []map[string]interface{}   `json:"defaultEnvironmentVariables,omitempty"  yaml:"defaultEnvironmentVariables,omitempty"`
	DefaultHeaders                       []DefaultHeader            `json:"defaultHeaders,omitempty"  yaml:"defaultHeaders,omitempty"`
	DefaultLabels                        []DefaultLabel             `json:"defaultLabels,omitempty"  yaml:"defaultLabels,omitempty"`
	DefaultLogParser                     *string                    `json:"defaultLogParser,omitempty"  yaml:"defaultLogParser,omitempty"`
	DefaultMonitoring                    *DefaultMonitoring         `json:"defaultMonitoring,omitempty"  yaml:"defaultMonitoring,omitempty"`
	DefaultProbes                        *DefaultProbes             `json:"defaultProbes,omitempty"  yaml:"defaultProbes,omitempty"`
	DefaultResources                     *DefaultResources          `json:"defaultResources,omitempty"  yaml:"defaultResources,omitempty"`
	DefaultSecrets                       []DefaultSecret            `json:"defaultSecrets,omitempty"  yaml:"defaultSecrets,omitempty"`
	DefaultTerminationGracePeriodSeconds *float64                   `json:"defaultTerminationGracePeriodSeconds,omitempty"  yaml:"defaultTerminationGracePeriodSeconds,omitempty"`
	Type                                 *string                    `json:"type,omitempty"  yaml:"type,omitempty"`
}

type DefaultContainerPort struct {
	From     interface{} `json:"from"  yaml:"from"`
	Name     string      `json:"name"  yaml:"name"`
	Protocol *string     `json:"protocol,omitempty"  yaml:"protocol,omitempty"`
	To       interface{} `json:"to"  yaml:"to"`
}

type DefaultCustomFilesConfig struct {
	FileName *string `json:"fileName,omitempty"  yaml:"fileName,omitempty"`
	FilePath *string `json:"filePath,omitempty"  yaml:"filePath,omitempty"`
	FileType *string `json:"fileType,omitempty"  yaml:"fileType,omitempty"`
	Ref      *string `json:"ref,omitempty"  yaml:"ref,omitempty"`
}

type DefaultAnnotation struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type DefaultConfigMap struct {
	Files          []DefaultFile `json:"files,omitempty"  yaml:"files,omitempty"`
	Link           *DefaultLink  `json:"link,omitempty"  yaml:"link,omitempty"`
	MountPath      *string       `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name           *string       `json:"name,omitempty"  yaml:"name,omitempty"`
	SubPaths       []string      `json:"subPaths,omitempty"  yaml:"subPaths,omitempty"`
	ViewAsReadOnly *bool         `json:"viewAsReadOnly,omitempty"  yaml:"viewAsReadOnly,omitempty"`
}

type DefaultFile struct {
	Content *string `json:"content,omitempty"  yaml:"content,omitempty"`
	Name    *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type DefaultLink struct {
	TargetSection *string `json:"targetSection,omitempty"  yaml:"targetSection,omitempty"`
}

type DefaultHeader struct {
	Description string `json:"description"  yaml:"description"`
	Name        string `json:"name"  yaml:"name"`
	Value       string `json:"value"  yaml:"value"`
}

type DefaultLabel struct {
	Description *string `json:"description,omitempty"  yaml:"description,omitempty"`
	Name        string  `json:"name"  yaml:"name"`
	ReadOnly    *bool   `json:"readOnly,omitempty"  yaml:"readOnly,omitempty"`
	Value       string  `json:"value"  yaml:"value"`
}

type DefaultMonitoring struct {
	Endpoints []DefaultEndpoint `json:"endpoints,omitempty"  yaml:"endpoints,omitempty"`
}

type DefaultEndpoint struct {
	Interval *string `json:"interval,omitempty"  yaml:"interval,omitempty"`
	Path     *string `json:"path,omitempty"  yaml:"path,omitempty"`
	Port     *string `json:"port,omitempty"  yaml:"port,omitempty"`
}

type DefaultProbes struct {
	Liveness  *DefaultLiveness  `json:"liveness,omitempty"  yaml:"liveness,omitempty"`
	Readiness *DefaultReadiness `json:"readiness,omitempty"  yaml:"readiness,omitempty"`
}

type DefaultLiveness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type DefaultReadiness struct {
	Cmd                 []string `json:"cmd,omitempty"  yaml:"cmd,omitempty"`
	FailureThreshold    *float64 `json:"failureThreshold,omitempty"  yaml:"failureThreshold,omitempty"`
	InitialDelaySeconds *float64 `json:"initialDelaySeconds,omitempty"  yaml:"initialDelaySeconds,omitempty"`
	Path                *string  `json:"path,omitempty"  yaml:"path,omitempty"`
	PeriodSeconds       *float64 `json:"periodSeconds,omitempty"  yaml:"periodSeconds,omitempty"`
	Port                *int64   `json:"port,omitempty"  yaml:"port,omitempty"`
	SuccessThreshold    *float64 `json:"successThreshold,omitempty"  yaml:"successThreshold,omitempty"`
	TimeoutSeconds      *float64 `json:"timeoutSeconds,omitempty"  yaml:"timeoutSeconds,omitempty"`
}

type DefaultResources struct {
	CPULimits    *DefaultCPULimits    `json:"cpuLimits,omitempty"  yaml:"cpuLimits,omitempty"`
	MemoryLimits *DefaultMemoryLimits `json:"memoryLimits,omitempty"  yaml:"memoryLimits,omitempty"`
}

type DefaultCPULimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type DefaultMemoryLimits struct {
	Max *string `json:"max,omitempty"  yaml:"max,omitempty"`
	Min *string `json:"min,omitempty"  yaml:"min,omitempty"`
}

type DefaultSecret struct {
	MountPath *string `json:"mountPath,omitempty"  yaml:"mountPath,omitempty"`
	Name      *string `json:"name,omitempty"  yaml:"name,omitempty"`
}

type Type string

const (
	ExternalLink Type = "externalLink"
	Markdown     Type = "markdown"
)
