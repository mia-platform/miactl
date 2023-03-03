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

package login

// import (
// 	"errors"
// 	"fmt"

// 	"github.com/mia-platform/miactl/old/factory"
// 	"github.com/mia-platform/miactl/old/sdk"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"
// )

// // NewLoginCmd create a new Login command
// func NewLoginCmd() *cobra.Command {
// 	var (
// 		baseURL         string
// 		username        string
// 		password        string
// 		providerID      string
// 		skipCertificate bool
// 		certificatePath string
// 	)

// 	cmd := &cobra.Command{
// 		Use:   "login",
// 		Short: "authenticate with console",
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			if baseURL = viper.GetString("apibaseurl"); baseURL == "" {
// 				return errors.New("API base URL not specified nor configured")
// 			}

// 			// set these flag only in case they are defined
// 			skipCertificate, _ = cmd.Flags().GetBool("insecure")
// 			certificatePath = viper.GetString("ca-cert")

// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			f, err := factory.FromContext(cmd.Context(), sdk.Options{
// 				APIBaseURL:            baseURL,
// 				SkipCertificate:       skipCertificate,
// 				AdditionalCertificate: certificatePath,
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			accessToken, err := f.MiaClient.Auth.Login(username, password, providerID)
// 			if err != nil {
// 				return err
// 			}

// 			// save current token for later commands
// 			viper.Set("apitoken", accessToken)
// 			if err = viper.WriteConfig(); err != nil {
// 				fmt.Println("error saving API token in the configuration")
// 				return err
// 			}

// 			fmt.Println("OK")
// 			return nil
// 		},
// 	}

// 	cmd.Flags().StringVar(&username, "username", "", "your user identifier")
// 	cmd.Flags().StringVar(&password, "password", "", "your secret password")
// 	cmd.Flags().StringVar(&providerID, "provider-id", "", "the authentication provider identifier")
// 	// Note: although this flag is defined as a persistent flag in the root command,
// 	// in order to be set during tests it must be defined also at command level
// 	cmd.Flags().BoolVar(&skipCertificate, "insecure", false, "whether to not check server certificate")

// 	cmd.MarkFlagRequired("username")
// 	cmd.MarkFlagRequired("password")
// 	cmd.MarkFlagRequired("provider-id")

// 	return cmd
// }
