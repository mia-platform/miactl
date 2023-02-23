package context

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSetContextCmd(opts *clioptions.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [flags]",
		Short: "update available contexts for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateContextMap(cmd, opts)
		},
	}
	return cmd
}

func updateContextMap(cmd *cobra.Command, opts *clioptions.RootOptions) error {
	contextMap := make(map[string]interface{})
	if viper.Get("contexts") != nil {
		contextMap = viper.Get("contexts").(map[string]interface{})
	}
	if contextMap[opts.Context] == nil {
		newContext := map[string]string{"apibaseurl": opts.APIBaseURL, "projectid": opts.ProjectID, "companyid": opts.CompanyID}
		contextMap[opts.Context] = newContext
	} else {
		oldContext := contextMap[opts.Context].(map[string]interface{})
		if opts.APIBaseURL != "https://console.cloud.mia-platform.eu" {
			oldContext["apibaseurl"] = opts.APIBaseURL
		}
		if opts.ProjectID != "" {
			oldContext["projectid"] = opts.ProjectID
		}
		if opts.CompanyID != "" {
			oldContext["companyid"] = opts.CompanyID
		}
	}
	viper.Set("contexts", contextMap)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("error saving API token in the configuration")
		return err
	}

	fmt.Println("OK")
	return nil
}
