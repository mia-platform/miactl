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

package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/old/sdk/auth"
	"github.com/mia-platform/miactl/old/sdk/deploy"
	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"
)

// Options struct define options to create the sdk client
type Options struct {
	APIKey                string
	APICookie             string
	APIBaseURL            string
	APIToken              string
	SkipCertificate       bool
	AdditionalCertificate string
	ProjectID             string
	CompanyID             string
	Context               string
}

// MiaClient is the client of the sdk to be used to communicate with Mia
// Platform Console api
type MiaClient struct {
	Projects deploy.IProjects
	Deploy   deploy.IDeploy
	Auth     auth.IAuth
}

// New returns the MiaSdkClient to be used to communicate to Mia Platform
// Console api.
func New(opts Options) (*MiaClient, error) {
	headers := jsonclient.Headers{}

	if opts.APIBaseURL == "" {
		return nil, fmt.Errorf("%w: client options are not correct", sdkErrors.ErrCreateClient)
	}

	// select auth method depending on given parameters
	if opts.APIToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIToken)
	} else if opts.APIKey != "" && opts.APICookie != "" {
		headers["cookie"] = opts.APICookie
		headers["client-key"] = opts.APIKey
	}

	clientOptions := jsonclient.Options{
		BaseURL: opts.APIBaseURL,
		Headers: headers,
	}

	customTransport, err := getCustomTransport(opts.SkipCertificate, opts.AdditionalCertificate)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrCreateClient, err)
	}
	clientOptions.HTTPClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: customTransport,
	}

	JSONClient, err := jsonclient.New(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrCreateClient, err)
	}

	return &MiaClient{
		Projects: &deploy.ProjectsClient{JSONClient: JSONClient},
		Deploy:   &deploy.Client{JSONClient: JSONClient},
		Auth:     &auth.Client{JSONClient: JSONClient},
	}, nil
}

func getCustomTransport(skipCertificate bool, additionalCertificate string) (*http.Transport, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipCertificate, //nolint:gosec
	}

	if additionalCertificate != "" {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			fmt.Println("error loading system cert pool - usign a new one")
			rootCAs = x509.NewCertPool()
		}

		cert, err := os.ReadFile(additionalCertificate)
		if err != nil {
			return nil, err
		}

		if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
			fmt.Println("no certs appended, using system certs only")
		}

		tlsConfig.RootCAs = rootCAs
	}

	customTransport.TLSClientConfig = tlsConfig

	return customTransport, nil
}
