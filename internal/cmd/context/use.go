package context

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var contextName string

func NewUseContextCmd(opts *clioptions.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use [flags]",
		Short: "update available contexts for miactl",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := contextLookUp(contextName); err != nil {
				return fmt.Errorf("error looking up the context in the config file: %w", err)
			}
			viper.Set("current-context", contextName)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("error updating the configuration: %w", err)
			}
			fmt.Println("OK")
			return nil
		},
	}

	cmd.Flags().StringVar(&contextName, "context-name", "default", "The name of the context to add or update")

	return cmd
}

func contextLookUp(contextName string) error {
	if viper.Get("contexts") == nil {
		return fmt.Errorf("no context specified in config file")
	}
	contextMap := viper.Get("contexts").(map[string]interface{})
	if contextMap[contextName] == nil {
		return fmt.Errorf("context %s does not exist", contextName)
	}
	return nil
}
