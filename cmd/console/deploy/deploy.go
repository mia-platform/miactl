package deploy

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type deployType string

const (
	smartDeploy deployType = "smart_deploy"
	deployAll              = "deploy_all"
)

type deployConfig struct {
	Environment         string
	Revision            string
	DeployAll           bool
	ForceDeployNoSemVer bool
}

type deployRequest struct {
	Environment             string     `json:"environment"`
	Revision                string     `json:"revision"`
	DeployType              deployType `json:"deployType"`
	ForceDeployWhenNoSemver bool       `json:"forceDeployWhenNoSemver"`
}

type deployResponse struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func NewDeployCmd() *cobra.Command {
	var (
		baseURL   string
		apiToken  string
		projectId string
	)

	cfg := deployConfig{}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy project",
		Long:  "trigger the deploy pipeline for selected project",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			baseURL = viper.GetString("apibaseurl")
			apiToken = viper.GetString("apitoken")
			projectId = viper.GetString("project")

			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}
			if apiToken == "" {
				return errors.New("missing API token - please login")
			}
			if projectId == "" {
				cmd.MarkFlagRequired("project")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r := renderer.New(cmd.OutOrStdout())

			deployData, err := deploy(baseURL, apiToken, projectId, &cfg)
			if err != nil {
				r.Error(err).Render()
				return nil
			}

			// store triggered pipelines details to enabling checking their status
			var pipelines Pipelines
			if err := viper.UnmarshalKey(triggeredPipelinesKey, &pipelines); err != nil {
				return err
			}

			viper.Set(triggeredPipelinesKey, append(pipelines, Pipeline{
				ProjectId:   projectId,
				PipelineId:  deployData.Id,
				Environment: cfg.Environment,
			}))
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			visualizeResponse(r, projectId, deployData)

			return nil
		},
	}

	cmd.Flags().StringVar(&cfg.Environment, "environment", "", "the environment where to deploy the project")
	cmd.Flags().StringVar(&cfg.Revision, "revision", "", "which version of your project should be released")
	cfg.DeployAll = *cmd.Flags().Bool("deploy-all", false, "deploy all the project services, regardless of whether they have been updated or not")
	cfg.ForceDeployNoSemVer = *cmd.Flags().Bool("force-no-semver", false, "whether to always deploy pods that do not follow semver")

	cmd.MarkFlagRequired("environment")
	cmd.MarkFlagRequired("revision")

	return cmd
}

func deploy(baseUrl, apiToken, projectId string, cfg *deployConfig) (deployResponse, error) {
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: baseUrl,
		Headers: jsonclient.Headers{
			"Authorization": fmt.Sprintf("Bearer %s", apiToken),
		},
	})
	if err != nil {
		return deployResponse{}, fmt.Errorf("error creating JSON client: %w", err)
	}

	data := deployRequest{
		Environment:             cfg.Environment,
		Revision:                cfg.Revision,
		DeployType:              smartDeploy,
		ForceDeployWhenNoSemver: cfg.ForceDeployNoSemVer,
	}

	if cfg.DeployAll == true {
		data.DeployType = deployAll
		data.ForceDeployWhenNoSemver = true
	}

	request, err := JSONClient.NewRequest(http.MethodPost, getDeployEndpoint(projectId), data)
	if err != nil {
		return deployResponse{}, fmt.Errorf("error creating deploy request: %w", err)
	}
	var response deployResponse

	rawRes, err := JSONClient.Do(request, &response)
	if err != nil {
		return deployResponse{}, fmt.Errorf("deploy error: %w", err)
	}
	rawRes.Body.Close()

	return response, nil
}

func getDeployEndpoint(projectId string) string {
	return fmt.Sprintf("deploy/projects/%s/trigger/pipeline/", projectId)
}

func visualizeResponse(r renderer.IRenderer, projectId string, rs deployResponse) {
	headers := []string{"Project Id", "Deploy Id", "View Pipeline"}
	table := r.Table(headers)
	table.Append([]string{projectId, strconv.FormatInt(int64(rs.Id), 10), rs.Url})
	table.Render()
}
