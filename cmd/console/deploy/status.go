package deploy

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewStatusCmd() *cobra.Command {
	var (
		baseURL     string
		apiToken    string
		projectId   string
		environment string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "verify status of deploy pipeline",
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
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

			pipelineId, err := strconv.Atoi(args[0])
			if err != nil {
				f.Renderer.Error(fmt.Errorf("unexpected pipeline id: %w", err)).Render()
				return nil
			}

			result, err := f.MiaClient.Deploy.GetDeployStatus(projectId, pipelineId, environment)
			if err != nil {
				f.Renderer.Error(err).Render()
				return nil
			}

			visualizeStatusResponse(f, projectId, result)

			switch result.Status {
			case sdk.Failed:
				return fmt.Errorf("Deploy pipeline failed")
			case sdk.Canceled:
				return fmt.Errorf("Deploy pipeline canceled")
			default:
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "the environment where the project has been deployed")

	return cmd
}

func visualizeStatusResponse(f *factory.Factory, projectId string, rs sdk.StatusResponse) {
	headers := []string{"Project Id", "Deploy Id", "Status"}
	table := f.Renderer.Table(headers)
	table.Append([]string{projectId, strconv.FormatInt(int64(rs.PipelineId), 10), rs.Status})
	table.Render()
}
