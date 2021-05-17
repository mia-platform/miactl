package deploy

import (
	"errors"
	"fmt"
	"time"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const checkDelay = 2 * time.Second
const endMessage = "all deploy pipelines triggered have completed"

func NewStatusCmd() *cobra.Command {
	var (
		baseURL  string
		apiToken string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "verify status of deploy pipeline",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			baseURL = viper.GetString("apibaseurl")
			apiToken = viper.GetString("apitoken")

			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}
			if apiToken == "" {
				return errors.New("missing API token - please login")
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

			var pipelines sdk.PipelinesConfig
			if err := readPipelines(&pipelines); err != nil {
				return err
			}

			if len(pipelines) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no deploy pipelines triggered found")
				return nil
			}

			tSleep := &sdk.TimeSleeper{Delay: checkDelay}
			lastEndedDeploy, err := f.MiaClient.Deploy.StatusMonitor(cmd.OutOrStdout(), &pipelines, tSleep)
			if err != nil {
				return err
			}

			if err := storePipelines(pipelines[lastEndedDeploy:]); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), endMessage)
			return nil
		},
	}

	return cmd
}
