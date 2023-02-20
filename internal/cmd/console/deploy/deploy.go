package deploy

import (
	"errors"
	"strconv"

	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/sdk"
	"github.com/mia-platform/miactl/old/sdk/deploy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDeployCmd() *cobra.Command {
	var (
		baseURL         string
		apiToken        string
		projectID       string
		skipCertificate bool
		certificatePath string
	)

	cfg := deploy.Config{}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy project",
		Long:  "trigger the deploy pipeline for selected project",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			baseURL = viper.GetString("apibaseurl")
			apiToken = viper.GetString("apitoken")
			projectID = viper.GetString("project")

			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}
			if apiToken == "" {
				return errors.New("missing API token - please login")
			}
			if projectID == "" {
				return cmd.MarkFlagRequired("project")
			}

			// set these flag only in case they are defined
			skipCertificate, _ = cmd.Flags().GetBool("insecure")
			certificatePath = viper.GetString("ca-cert")

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := factory.FromContext(cmd.Context(), sdk.Options{
				APIBaseURL:            baseURL,
				APIToken:              apiToken,
				SkipCertificate:       skipCertificate,
				AdditionalCertificate: certificatePath,
			})
			if err != nil {
				return err
			}

			deployData, err := f.MiaClient.Deploy.Trigger(projectID, cfg)
			if err != nil {
				return err
			}

			visualizeResponse(f, projectID, deployData)

			return nil
		},
	}

	cmd.Flags().StringVar(&cfg.Environment, "environment", "", "the environment where to deploy the project")
	cmd.Flags().StringVar(&cfg.Revision, "revision", "", "which version of your project should be released")
	cmd.Flags().BoolVar(&cfg.DeployAll, "deploy-all", false, "deploy all the project services, regardless of whether they have been updated or not")
	cmd.Flags().BoolVar(&cfg.ForceDeployNoSemVer, "force-no-semver", false, "whether to always deploy pods that do not follow semver")
	// Note: although this flag is defined as a persistent flag in the root command,
	// in order to be set during tests it must be defined also at command level
	cmd.Flags().BoolVar(&skipCertificate, "insecure", false, "whether to not check server certificate")

	cmd.MarkFlagRequired("environment")
	cmd.MarkFlagRequired("revision")

	// subcommands
	cmd.AddCommand(NewStatusCmd())

	return cmd
}

func visualizeResponse(f *factory.Factory, projectID string, rs deploy.Response) {
	headers := []string{"Project Id", "Deploy Id", "View Pipeline"}
	table := f.Renderer.Table(headers)
	table.Append([]string{projectID, strconv.FormatInt(int64(rs.ID), 10), rs.URL})
	table.Render()
}
