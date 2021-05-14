package deploy

import (
	"errors"

	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk/deploy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDeployCmd() *cobra.Command {
	var (
		baseURL   string
		apiToken  string
		projectId string
	)

	cfg := deploy.DeployConfig{}

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

			deployData, err := deploy.Trigger(baseURL, apiToken, projectId, cfg)
			if err != nil {
				r.Error(err).Render()
				return nil
			}

			var pipelines deploy.Pipelines
			if err := deploy.ReadPipelines(&pipelines); err != nil {
				return err
			}

			if err := deploy.StorePipelines(pipelines, deploy.Pipeline{
				ProjectId:   projectId,
				PipelineId:  deployData.Id,
				Environment: cfg.Environment,
			}); err != nil {
				return err
			}

			deploy.VisualizeResponse(r, projectId, deployData)

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
