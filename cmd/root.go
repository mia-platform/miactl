package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/cmd/console"
	"github.com/mia-platform/miactl/cmd/login"
	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/sdk"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cfgFileName = ".miaplatformctl.yaml"

var (
	cfgFile   string
	projectID string
	opts      = sdk.Options{}
	verbose   bool
)

// NewRootCmd creates a new root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "miactl",
		SilenceUsage: true,
		/* SilenceErrors must be set to true, since a custom error visualization
		logic is implemented in the execute command below.
		On the contrary, errors are visualized twice or more */
		SilenceErrors: true,
	}
	setRootPersistentFlag(rootCmd)

	// add sub command to root command
	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(login.NewLoginCmd())
	rootCmd.AddCommand(console.NewConsoleCmd())

	rootCmd.AddCommand(newCompletionCmd(rootCmd))
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCmd()
	ctx := factory.WithValue(context.Background(), rootCmd.OutOrStdout())
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
	rootCmd.PersistentFlags().StringVar(&opts.APIToken, "apiToken", "", "api access token")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "specify desired project ID")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "whether to output details in verbose mode")
	rootCmd.PersistentFlags().BoolVar(&opts.SkipCertificate, "insecure", false, "whether to not check server certificate")
	rootCmd.PersistentFlags().StringVar(
		&opts.AdditionalCertificate,
		"ca-cert",
		"",
		"file path to additional CA certificate, which can be employed to verify server certificate",
	)

	rootCmd.MarkFlagRequired("apiBaseUrl")

	viper.BindPFlag("apibaseurl", rootCmd.PersistentFlags().Lookup("apiBaseUrl"))
	viper.BindPFlag("apitoken", rootCmd.PersistentFlags().Lookup("apiToken"))
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("ca-cert", rootCmd.PersistentFlags().Lookup("ca-cert"))
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

		// create a default config file if it does not exist
		if err := viper.SafeWriteConfig(); err != nil {
			viper.SafeWriteConfigAs(path.Join(home, cfgFileName))
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error loading file:", viper.ConfigFileUsed())
	}
	if verbose {
		fmt.Printf("Using config file: %s\n\n", viper.ConfigFileUsed())
	}
}
