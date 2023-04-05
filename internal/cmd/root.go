// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/console/deploy"
	miacontext "github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/project"
	"github.com/mia-platform/miactl/old/factory"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfgDir         = ".config/miactl"
	cfgFileName    = "config"
	credentialsDir = "credentials"
)

var (
	cfgFile string
	verbose bool
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

	options := clioptions.NewCLIOptions()
	options.AddRootFlags(rootCmd)

	// add sub command to root command
	rootCmd.AddCommand(project.NewProjectCmd(options))
	rootCmd.AddCommand(deploy.NewDeployCmd(options))
	rootCmd.AddCommand(miacontext.NewContextCmd(options))
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

		cfgPath := path.Join(home, cfgDir)

		// Search config in home directory with name ".miaplatformctl" (without extension).
		viper.SetConfigName(cfgFileName)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(cfgPath)

		// create a default config file if it does not exist
		if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := viper.SafeWriteConfigAs(path.Join(cfgPath, cfgFileName)); err != nil && verbose {
			fmt.Println(err)
		}

		credPath := path.Join(cfgPath, credentialsDir)

		// create a default config file if it does not exist
		if err := os.MkdirAll(credPath, os.ModePerm); err != nil {
			fmt.Println(err)
			os.Exit(1)
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
