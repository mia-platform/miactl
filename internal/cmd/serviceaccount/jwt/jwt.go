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

package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/resources"
	"github.com/spf13/cobra"
)

const (
	rsaKeyBytes                            = 4096
	companyServiceAccountsEndpointTemplate = "/api/companies/%s/service-accounts"
	defaultJSONType                        = "service_account"
	defaultKeyID                           = "miactl"
)

func ServiceAccountCmd(options *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jwt SERVICEACCOUNT [flags]",
		Short: "Create a new jwt authentication service account",
		Long: `Create a new jwt authentication service account in the provided company.

You can create a service account with the same or lower role than the role that
the current authentication has. The role company-owner can be used only when the
service account is created on the company.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceAccountName := args[0]
			restConfig, err := options.ToRESTConfig()
			cobra.CheckErr(err)
			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)
			credentials, err := createJWTServiceAccount(client, serviceAccountName, restConfig.CompanyID, resources.ServiceAccountRole(options.ServiceAccountRole))
			if err != nil {
				return err
			}

			return saveCredentialsIfNeeded(credentials, options.OutputPath, cmd.OutOrStdout())
		},
	}

	// add cmd flags
	options.AddJWTServiceAccountFlags(cmd.Flags())
	err := cmd.RegisterFlagCompletionFunc("role", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			resources.ServiceAccountRoleGuest.String(),
			resources.ServiceAccountRoleReporter.String(),
			resources.ServiceAccountRoleDeveloper.String(),
			resources.ServiceAccountRoleMaintainer.String(),
			resources.ServiceAccountRoleProjectAdmin.String(),
			resources.ServiceAccountRoleCompanyOwner.String(),
		}, cobra.ShellCompDirectiveDefault
	})

	if err != nil {
		// we panic here because if we reach here, something nasty is happenign in flag autocomplete registration
		panic(err)
	}

	err = cmd.MarkFlagDirname("output")
	if err != nil {
		// we panic here because if we reach here, something nasty is happenign in flag autocomplete registration
		panic(err)
	}

	return cmd
}

func createJWTServiceAccount(client *client.APIClient, name, companyID string, role resources.ServiceAccountRole) (*resources.JWTServiceAccountJSON, error) {
	if !resources.IsValidServiceAccountRole(role) {
		return nil, fmt.Errorf("invalid service account role %s", role)
	}

	if len(companyID) == 0 {
		return nil, fmt.Errorf("company id is required, please set it via flag or context")
	}

	key, err := generateRSAKey()
	if err != nil {
		return nil, err
	}

	payload := requestFromKey(name, role, key)
	body, err := resources.EncodeResourceToJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	resp, err := client.
		Post().
		SetAPIPath(fmt.Sprintf(companyServiceAccountsEndpointTemplate, companyID)).
		SetBody(body).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	if err := resp.Error(); err != nil {
		return nil, err
	}

	response := new(resources.ServiceAccount)
	if err := resp.ParseResponse(response); err != nil {
		return nil, err
	}

	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	pemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pkcs8,
		},
	)

	return &resources.JWTServiceAccountJSON{
		Type:           defaultJSONType,
		KeyID:          defaultKeyID,
		PrivateKeyData: base64.StdEncoding.EncodeToString(pemData),
		ClientID:       response.ClientID,
	}, nil
}

func saveCredentialsIfNeeded(credentials *resources.JWTServiceAccountJSON, outputPath string, stdout io.Writer) error {
	var encoder *json.Encoder
	var fileDest *os.File
	if len(outputPath) > 0 {
		var err error
		fileDest, err = os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Service account created, credentials saved in %s\n", outputPath)
		encoder = json.NewEncoder(fileDest)
	} else {
		fmt.Fprintln(stdout, "Service account created, save the following json for later uses:")
		encoder = json.NewEncoder(stdout)
	}

	defer func() {
		if fileDest != nil {
			fileDest.Close()
		}
	}()

	encoder.SetIndent("", "	")
	return encoder.Encode(credentials)
}

func generateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeyBytes)
}

func requestFromKey(name string, role resources.ServiceAccountRole, key *rsa.PrivateKey) *resources.ServiceAccountRequest {
	encoder := base64.RawURLEncoding
	modulus, exponent := rsaPublicKeyToBytes(&key.PublicKey)

	return &resources.ServiceAccountRequest{
		Name: name,
		Role: role,
		Type: resources.ServiceAccountJWT,
		PublicKey: resources.PublicKey{
			Use:       "sig",
			Type:      "RSA",
			Algorithm: "RSA256",
			KeyID:     defaultKeyID,
			Modulus:   encoder.EncodeToString(modulus),
			Exponent:  encoder.EncodeToString(exponent),
		},
	}
}

// rsaPublicKeyToBytes take an RSA PublicKey struct as inpunt and return two
// bytes array, that follows the  https://www.rfc-editor.org/rfc/rfc7518#section-6.3.1
// specification, needed by a JWK
func rsaPublicKeyToBytes(key *rsa.PublicKey) ([]byte, []byte) {
	modulus := key.N.Bytes()

	// convert exponent in 8 byte and then truncate until the first byte set to 1
	exponentData := make([]byte, 8)
	binary.BigEndian.PutUint64(exponentData, uint64(key.E))
	i := 0
	var emptyByte byte = 0x0
	for ; i < len(exponentData); i++ {
		if exponentData[i] != emptyByte {
			break
		}
	}
	return modulus, exponentData[i:]
}
