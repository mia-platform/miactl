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

package project

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	importCmdUsage = "import"
	importCmdShort = "[beta] - Import kubernetes resources"
	importCmdLong  = `[beta] - Import kubernetes resources in a Mia-Platform Console project.`

	configurationEndpointTemplate = "/api/backend/projects/%s/revisions/%s/configuration"
	convertEndpointTemplate       = "/api/projects/%s/configurations/from-raw"
)

type convertResourcesBody struct {
	Runtime   string `json:"runtime"`
	Resources string `json:"resources"`
}

type convertedResourcesBody struct {
	Services        map[string]any `json:"services"`
	ConfigMaps      map[string]any `json:"configMaps"`
	Secrets         map[string]any `json:"secrets"`
	ServiceAccounts map[string]any `json:"serviceAccounts"`
	Errors          []string       `json:"errors"`
}

// ImportCmd return a cobra command for listing projects
func ImportCmd(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   importCmdUsage,
		Short: importCmdShort,
		Long:  importCmdLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			return importResources(cmd.Context(), client, restConfig.ProjectID, o.Revision, o.InputFilePath, cmd.ErrOrStderr())
		},
	}

	// add cmd flags
	flags := cmd.Flags()
	o.AddImportFlags(flags)

	return cmd
}

func importResources(ctx context.Context, client *client.APIClient, projectID, revision, path string, writer io.Writer) error {
	if len(projectID) == 0 {
		return errors.New("missing project id, please set one with the flag or context")
	}

	if len(revision) == 0 {
		return errors.New("missing revision, please set one with the revision flag")
	}

	if len(path) == 0 {
		return errors.New("missing file path, please set on with the filename flag")
	}

	endpoint := fmt.Sprintf(configurationEndpointTemplate, projectID, revision)
	return saveConfiguration(ctx, client, endpoint, path, projectID, writer)
}

func saveConfiguration(ctx context.Context, client *client.APIClient, endpoint, path, projectID string, writer io.Writer) error {
	convertedResources, err := convertResources(ctx, client, path, projectID)
	if err != nil {
		return err
	}

	if len(convertedResources.Services) == 0 && len(convertedResources.ConfigMaps) == 0 && len(convertedResources.Secrets) == 0 {
		fmt.Fprintln(writer, "No valid resources found to import")
		return nil
	}

	response, err := client.
		Get().
		APIPath(endpoint).
		Do(ctx)

	if err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	projectConfig := make(map[string]any, 0)
	if err := response.ParseResponse(&projectConfig); err != nil {
		return fmt.Errorf("cannot parse project configuration: %w", err)
	}

	newConfig, err := postBodyFromConfiguration(projectConfig, convertedResources)
	if err != nil {
		return err
	}

	body, err := resources.EncodeResourceToJSON(newConfig)
	if err != nil {
		return fmt.Errorf("cannot encode project configuration: %w", err)
	}

	response, err = client.
		Post().
		APIPath(endpoint).
		Body(body).
		Do(ctx)
	if err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	if len(convertedResources.Errors) > 0 {
		fmt.Fprintln(writer, "Configuration imported successfully with warnings:")
		for _, err := range convertedResources.Errors {
			fmt.Fprintf(writer, "\t- %s\n", err)
		}
	} else {
		fmt.Fprintln(writer, "Configuration imported successfully")
	}

	return nil
}

func convertResources(ctx context.Context, client *client.APIClient, path, projectID string) (*convertedResourcesBody, error) {
	rawResources, err := readRawResources(path)
	if err != nil {
		return nil, err
	}

	return getConvertedResources(ctx, client, projectID, rawResources)
}

func getConvertedResources(ctx context.Context, client *client.APIClient, projectID string, res []map[string]any) (*convertedResourcesBody, error) {
	buffer := bytes.NewBuffer([]byte{})
	if err := writeResources(buffer, res); err != nil {
		return nil, err
	}

	body, err := resources.EncodeResourceToJSON(convertResourcesBody{
		Runtime:   "kubernetes",
		Resources: buffer.String(),
	})

	if err != nil {
		return nil, err
	}

	response, err := client.
		Post().
		APIPath(fmt.Sprintf(convertEndpointTemplate, projectID)).
		Body(body).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	if response.StatusCode() == http.StatusNotFound {
		return nil, errors.New(`this command is only available on Mia-Platform Console Preview environment`)
	}

	if err := response.Error(); err != nil {
		return nil, err
	}

	convertedResources := new(convertedResourcesBody)
	if err := response.ParseResponse(convertedResources); err != nil {
		return nil, err
	}

	return convertedResources, nil
}

func postBodyFromConfiguration(config map[string]any, newResources *convertedResourcesBody) (map[string]any, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	postConfig := make(map[string]any, 0)

	delete(config, "committedDate")
	delete(config, "lastCommitAuthor")
	delete(config, "platformVersion")

	postConfig["fastDataConfig"] = config["fastDataConfig"]
	postConfig["microfrontendPluginsConfig"] = config["microfrontendPluginsConfig"]
	if postConfig["microfrontendPluginsConfig"] == nil {
		postConfig["microfrontendPluginsConfig"] = make(map[string]any, 0)
	}
	postConfig["extensionsConfig"] = config["extensionsConfig"]
	if postConfig["extensionsConfig"] == nil {
		postConfig["extensionsConfig"] = map[string]any{
			"files": make(map[string]any, 0),
		}
	}

	delete(config, "fastDataConfig")
	delete(config, "microfrontendPluginsConfig")
	delete(config, "extensionsConfig")

	config["services"] = newResources.Services
	config["configMaps"] = newResources.ConfigMaps
	config["serviceSecrets"] = newResources.Secrets
	if len(newResources.ServiceAccounts) != 0 {
		config["serviceAccounts"] = newResources.ServiceAccounts
	}

	postConfig["config"] = config
	postConfig["previousSave"] = config["commitId"]
	postConfig["title"] = "[CLI] Import resource from kubernetes"
	postConfig["deletedElements"] = make(map[string]any, 0)
	return postConfig, nil
}

func readRawResources(path string) ([]map[string]any, error) {
	reader := &kio.LocalPackageReader{
		PackagePath:           path,
		OmitReaderAnnotations: true,
	}

	var objs []map[string]any

	pipeline := kio.Pipeline{
		Inputs:  []kio.Reader{reader},
		Filters: []kio.Filter{filters.StripCommentsFilter{}, &filters.IsLocalConfig{}},
		Outputs: []kio.Writer{kio.WriterFunc(func(nodes []*yaml.RNode) error {
			for _, node := range nodes {
				data, err := node.MarshalJSON()
				if err != nil {
					return err
				}

				var object map[string]any
				if err = json.Unmarshal(data, &object); err != nil {
					return err
				}

				objs = append(objs, object)
			}

			return nil
		})},
	}

	if err := pipeline.Execute(); err != nil {
		return nil, err
	}

	return objs, nil
}

func writeResources(writer io.Writer, resources []map[string]any) error {
	firstObjs := true
	for _, resource := range resources {
		yml, err := yaml.Marshal(resource)
		if err != nil {
			return err
		}
		if firstObjs {
			firstObjs = false
		} else {
			if _, err := writer.Write([]byte("---\n")); err != nil {
				return err
			}
		}
		if _, err := writer.Write(yml); err != nil {
			return err
		}
	}

	return nil
}

func validateConfig(config map[string]any) error {
	if config["services"] != nil {
		services, ok := config["services"].(map[string]any)
		if ok && len(services) > 0 {
			return errors.New("cannot import services in a non empty project")
		}
	}

	return nil
}
