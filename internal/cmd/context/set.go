package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSetContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [flags]",
		Short: "update available contexts for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateContextMap(cmd)
		},
	}
	return cmd
}

func updateContextMap(cmd *cobra.Command) error {
	o := getOptions(cmd)
	contextMap := make(map[string]interface{})
	if viper.Get("contexts") != nil {
		contextMap = viper.Get("contexts").(map[string]interface{})
	}
	if contextMap[o.name] == nil {
		newContext := map[string]string{"apibaseurl": o.endpoint, "projectid": o.projectID, "companyid": o.companyID}
		contextMap[o.name] = newContext
	} else {
		oldContext := contextMap[o.name].(map[string]interface{})
		if o.endpoint != "https://console.cloud.mia-platform.eu" {
			oldContext["apibaseurl"] = o.endpoint
		}
		if o.projectID != "" {
			oldContext["projectid"] = o.projectID
		}
		if o.companyID != "" {
			oldContext["companyid"] = o.companyID
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
