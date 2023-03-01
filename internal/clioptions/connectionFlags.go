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

package clioptions

import "github.com/spf13/cobra"

type ConnectionOptions struct {
	APIKey                string
	APICookie             string
	APIToken              string
	SkipCertificate       bool
	AdditionalCertificate string
	Context               string
}

func (f *ConnectionOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&f.APIKey, "apiKey", "", "API Key")
	cmd.PersistentFlags().StringVar(&f.APICookie, "apiCookie", "", "api cookie sid")
	cmd.PersistentFlags().StringVar(&f.APIToken, "apiToken", "", "api access token")
	cmd.PersistentFlags().StringVar(&f.Context, "context", "", "The name of the context to use")
	cmd.PersistentFlags().BoolVar(&f.SkipCertificate, "insecure", false, "whether to not check server certificate")
	cmd.PersistentFlags().StringVar(
		&f.AdditionalCertificate,
		"ca-cert",
		"",
		"file path to additional CA certificate, which can be employed to verify server certificate",
	)
}

func NewConnectionOptions() *ConnectionOptions {
	return &ConnectionOptions{}
}
