package deploy

import (
	"errors"
	"strconv"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/sdk"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const triggeredPipelinesKey = "triggered-pipelines"

// pipelineConfig represents a single triggered deploy pipeline in the config file.
type pipelineConfig struct {
	ProjectId   string `yaml:"projectid" mapstructure:"projectid"`
	PipelineId  int    `yaml:"pipelineid" mapstructure:"pipelineid"`
	Environment string `yaml:"environment" mapstructure:"environment"`
}

// pipelinesConfig represents a list in the config of all the pipelines
// that have been triggered but not checked their status.
type pipelinesConfig []pipelineConfig

func NewDeployCmd() *cobra.Command {
	var (
		baseURL   string
		apiToken  string
		projectId string
	)

	cfg := sdk.DeployConfig{}

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
				return cmd.MarkFlagRequired("project")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := factory.FromContext(cmd.Context(), sdk.Options{
				APIBaseURL: baseURL,
				APIToken:   apiToken,
			})
			if err != nil {
				return err
			}

			deployData, err := f.MiaClient.Deploy.Trigger(projectId, cfg)
			if err != nil {
				f.Renderer.Error(err).Render()
				return nil
			}

			visualizeResponse(f, projectId, deployData)

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

func visualizeResponse(f *factory.Factory, projectId string, rs sdk.DeployResponse) {
	headers := []string{"Project Id", "Deploy Id", "View Pipeline"}
	table := f.Renderer.Table(headers)
	table.Append([]string{projectId, strconv.FormatInt(int64(rs.Id), 10), rs.Url})
	table.Render()
}
