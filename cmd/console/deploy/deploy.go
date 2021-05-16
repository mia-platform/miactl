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

			var pipelines sdk.PipelinesConfig
			if err := readPipelines(&pipelines); err != nil {
				return err
			}

			if err := storePipelines(pipelines, sdk.PipelineConfig{
				ProjectId:   projectId,
				PipelineId:  deployData.Id,
				Environment: cfg.Environment,
			}); err != nil {
				return err
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

func storePipelines(ps sdk.PipelinesConfig, p sdk.PipelineConfig) error {
	viper.Set(triggeredPipelinesKey, append(ps, p))
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// ReadPipelines store triggered pipelines details to enabling checking their status
func readPipelines(p *sdk.PipelinesConfig) error {
	if err := viper.UnmarshalKey(triggeredPipelinesKey, p); err != nil {
		return err
	}

	return nil
}
