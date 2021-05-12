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

type deployConfig struct {
	Environment         string
	Revision            string
	SmartDeploy         bool
	ForceDeployNoSemVer bool
}

type deployRequest struct {
	Environment             string `json:"environment"`
	Revision                string `json:"revision"`
	DeployType              string `json:"deployType"`
	ForceDeployWhenNoSemver bool   `json:"forceDeployWhenNoSemver,omitempty"`
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

			if apiToken == "" {
				return errors.New("missing API token - please login")
			}
			if projectId == "" {
				cmd.MarkFlagRequired("project")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}
			if projectId == "" {
				return errors.New("project id not specified nor configured")
			}
			r := renderer.New(cmd.OutOrStdout())

			deployData, err := deploy(baseURL, apiToken, projectId, &cfg)
			if err != nil {
				r.Error(err).Render()
				return nil
			}

			// save pipeline id to simplify getting its state
			viper.Set("project-deploy-pipeline", deployData.Id)
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			visualizeResponse(r, projectId, deployData)

			return nil
		},
	}

	cmd.Flags().StringVar(&cfg.Environment, "environment", "", "the environment where to deploy the project")
	cmd.Flags().StringVar(&cfg.Revision, "revision", "", "which version of your project should be released")
	cfg.SmartDeploy = *cmd.Flags().Bool("smart-deploy", false, "enable smart-deploy feature, which deploys only updated resources")
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
		ForceDeployWhenNoSemver: cfg.ForceDeployNoSemVer,
	}

	if cfg.SmartDeploy == true {
		data.DeployType = "smart_deploy"
	} else {
		data.DeployType = "deploy_all"
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
	defer rawRes.Body.Close()

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
