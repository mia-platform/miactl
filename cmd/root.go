package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mia-platform/miactl/sdk"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	opts    = sdk.Options{}
)

// NewRootCmd creates a new root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "miactl",
	}
	setRootPersistentFlag(rootCmd)

	// add sub command to root command
	rootCmd.AddCommand(newGetCmd())

	rootCmd.AddCommand(newCompletionCmd(rootCmd))
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCmd()
	ctx := WithFactoryValue(context.Background(), rootCmd.OutOrStdout())
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func setRootPersistentFlag(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.miaplatformctl.yaml)")
	rootCmd.PersistentFlags().StringVar(&opts.APIKey, "apiKey", "", "API Key")
	rootCmd.PersistentFlags().StringVar(&opts.APICookie, "apiCookie", "", "api cookie sid")
	rootCmd.PersistentFlags().StringVar(&opts.APIBaseURL, "apiBaseUrl", "", "api base url")

	rootCmd.MarkFlagRequired("apiKey")
	rootCmd.MarkFlagRequired("apiCookie")
	rootCmd.MarkFlagRequired("apiBaseUrl")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".miaplatformctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".miaplatformctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
