package sdk

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
)

// Options struct define options to create the sdk client
type Options struct {
	APIKey          string
	APICookie       string
	APIBaseURL      string
	APIToken        string
	SkipCertificate bool
}

// IProjects expose the projects client interface
type IProjects interface {
	Get() (Projects, error)
}

// DeployHistoryQuery wraps query filters for project deployments.
type DeployHistoryQuery struct {
	ProjectID string
}

// IDeploy is a client interface used to interact with deployment pipelines.
type IDeploy interface {
	GetHistory(DeployHistoryQuery) ([]DeployItem, error)
	Trigger(string, DeployConfig) (DeployResponse, error)
	GetDeployStatus(string, int, string) (StatusResponse, error)
}

// MiaClient is the client of the sdk to be used to communicate with Mia
// Platform Console api
type MiaClient struct {
	Projects IProjects
	Deploy   IDeploy
}

var (
	// ErrGeneric is a generic error
	ErrGeneric = errors.New("Something went wrong")
	// ErrHTTP is an http error
	ErrHTTP = jsonclient.ErrHTTP
	// ErrCreateClient is the error creating MiaSdkClient
	ErrCreateClient = errors.New("Error creating sdk client")
	// ErrProjectNotFound is the error returned when specified
	// project cannot be found.
	ErrProjectNotFound = errors.New("Project not found")
)

// New returns the MiaSdkClient to be used to communicate to Mia Platform
// Console api.
func New(opts Options) (*MiaClient, error) {
	headers := jsonclient.Headers{}

	if opts.APIBaseURL == "" || (opts.APIToken == "" && (opts.APIKey == "" || opts.APICookie == "")) {
		return nil, fmt.Errorf("%w: client options are not correct", ErrCreateClient)
	}

	// select auth method depending on given parameters
	if opts.APIToken != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", opts.APIToken)
	} else {
		headers["cookie"] = opts.APICookie
		headers["client-key"] = opts.APIKey
	}

	clientOptions := jsonclient.Options{
		BaseURL: opts.APIBaseURL,
		Headers: headers,
	}

	if opts.SkipCertificate {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		clientOptions.HTTPClient = &http.Client{Transport: customTransport}
	}

	JSONClient, err := jsonclient.New(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateClient, err)
	}

	return &MiaClient{
		Projects: &ProjectsClient{JSONClient: JSONClient},
		Deploy:   &DeployClient{JSONClient: JSONClient},
	}, nil
}
