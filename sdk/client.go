package sdk

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/sdk/auth"
	"github.com/mia-platform/miactl/sdk/deploy"
	sdkErrors "github.com/mia-platform/miactl/sdk/errors"
)

// Options struct define options to create the sdk client
type Options struct {
	APIKey          string
	APICookie       string
	APIBaseURL      string
	APIToken        string
	SkipCertificate bool
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

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: opts.SkipCertificate,
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
		Deploy:   &deploy.DeployClient{JSONClient: JSONClient},
		Auth:     &auth.AuthClient{JSONClient: JSONClient},
	}, nil
}
