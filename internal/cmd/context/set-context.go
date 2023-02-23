package context

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type setContextFlags struct {
	RootOptions *clioptions.RootOptions
}

type setContextOptions struct {
	context   string
	endpoint  string
	projectID string
	companyID string
}

func NewSetContextCommand() *cobra.Command {
	flags := newSetContextFlags()

	cmd := &cobra.Command{
		Use:   "set-context [flags]",
		Short: "set current context for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {
			o := getOptions(cmd)

			contextMap := make(map[string]interface{})
			if viper.Get("contexts") != nil {
				contextMap = viper.Get("contexts").(map[string]interface{})
			}
			if contextMap[o.context] == nil {
				newContext := map[string]string{"apibaseurl": o.endpoint, "projectid": o.projectID, "companyid": o.companyID}
				contextMap[o.context] = newContext
			} else {
				oldContext := contextMap[o.context].(map[string]interface{})
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
		},
	}

	flags.addFlags(cmd)

	return cmd
}

func newSetContextFlags() *setContextFlags {
	return &setContextFlags{
		RootOptions: clioptions.NewRootOptions(),
	}
}

func (f *setContextFlags) addFlags(c *cobra.Command) {
	//root flags
	f.RootOptions.AddFlags(c)
}

func getOptions(c *cobra.Command) *setContextOptions {
	return &setContextOptions{
		context:   clioptions.GetFlagString(c, "context"),
		endpoint:  clioptions.GetFlagString(c, "endpoint"),
		projectID: clioptions.GetFlagString(c, "project-id"),
		companyID: clioptions.GetFlagString(c, "company-id"),
	}
}
