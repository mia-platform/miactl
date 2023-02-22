package context

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var contextName string

func NewSetContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-context [flags]",
		Short: "set current context for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {

			contextMap := make(map[string]interface{})
			if viper.Get("contexts") != nil {
				contextMap = viper.Get("contexts").(map[string]interface{})
			}
			if contextMap[contextName] == nil {
				newContext := map[string]string{"apibaseurl": clioptions.Opts.APIBaseURL, "projectid": clioptions.Opts.ProjectID, "companyid": clioptions.Opts.CompanyID}
				contextMap[contextName] = newContext
			} else {
				oldContext := contextMap[contextName].(map[string]string)
				if clioptions.Opts.APIBaseURL != "https://console.cloud.mia-platform.eu" {
					oldContext["apibaseurl"] = clioptions.Opts.APIBaseURL
				}
				if clioptions.Opts.ProjectID != "" {
					oldContext["projectid"] = clioptions.Opts.APIBaseURL
				}
				if clioptions.Opts.CompanyID != "" {
					oldContext["companyid"] = clioptions.Opts.APIBaseURL
				}
			}
			viper.Set("contexts", contextMap)
			if err := viper.WriteConfig(); err != nil {
				fmt.Println("error saving API token in the configuration")
				return err
			}

			fmt.Println("OK")
			return nil
		},
	}

	cmd.Flags().StringVar(&contextName, "context-name", "default", "The name of the context to add")

	return cmd
}
