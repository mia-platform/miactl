package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSetContextCommand() *cobra.Command {
	var (
		consoleUrl string
		companyID  string
		projectID  string
	)

	cmd := &cobra.Command{
		Use:   "set-context [flags]",
		Short: "set current context for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("apibaseurl", consoleUrl)
			viper.Set("projectid", projectID)
			viper.Set("companyid", companyID)
			if err := viper.WriteConfig(); err != nil {
				fmt.Println("error saving API token in the configuration")
				return err
			}

			fmt.Println("OK")
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "The ID of the project")
	cmd.Flags().StringVar(&companyID, "company-id", "", "The ID of the company")
	cmd.Flags().StringVar(&consoleUrl, "console-url", "", "The URL of the console")
	// Note: although this flag is defined as a persistent flag in the root command,
	// in order to be set during tests it must be defined also at command level

	cmd.MarkFlagRequired("company-id")

	return cmd
}
