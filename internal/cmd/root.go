package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/internal/cmd/console"
	miacontext "github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mia-platform/miactl/old/factory"
	"github.com/mia-platform/miactl/old/sdk"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cfgDir = ".config/miactl"
const cfgFileName = "config"

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
	rootCmd.AddCommand(miacontext.NewSetContextCommand())

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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/miactl/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&opts.ProjectID, "projectID", "", "The ID of the project")
	rootCmd.PersistentFlags().StringVar(&opts.CompanyID, "companyID", "", "The ID of the company")
	rootCmd.PersistentFlags().StringVar(&opts.APIKey, "apiKey", "", "API Key")
	rootCmd.PersistentFlags().StringVar(&opts.APICookie, "apiCookie", "", "api cookie sid")
	rootCmd.PersistentFlags().StringVar(&opts.APIBaseURL, "apiBaseUrl", "", "api base url")
	rootCmd.PersistentFlags().StringVar(&opts.APIToken, "apiToken", "", "api access token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "whether to output details in verbose mode")
	rootCmd.PersistentFlags().BoolVar(&opts.SkipCertificate, "insecure", false, "whether to not check server certificate")
	rootCmd.PersistentFlags().StringVar(
		&opts.AdditionalCertificate,
		"ca-cert",
		"",
		"file path to additional CA certificate, which can be employed to verify server certificate",
	)

	rootCmd.MarkFlagRequired("apiBaseUrl")

	viper.BindPFlag("projectID", rootCmd.PersistentFlags().Lookup("projectID"))
	viper.BindPFlag("companyID", rootCmd.PersistentFlags().Lookup("companyID"))
	viper.BindPFlag("apibaseurl", rootCmd.PersistentFlags().Lookup("apiBaseUrl"))
	viper.BindPFlag("apitoken", rootCmd.PersistentFlags().Lookup("apiToken"))
	viper.BindPFlag("ca-cert", rootCmd.PersistentFlags().Lookup("ca-cert"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigType("yaml")

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

		cfgPath := path.Join(home, cfgDir)

		// Search config in home directory with name ".miaplatformctl" (without extension).
		viper.AddConfigPath(cfgPath)
		viper.SetConfigName(cfgFileName)

		// create a default config file if it does not exist
		if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := viper.SafeWriteConfig(); err != nil {
			err = viper.SafeWriteConfigAs(path.Join(cfgPath, cfgFileName))
			fmt.Println(err)
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
