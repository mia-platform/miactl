package sdk

import (
	"errors"
	"fmt"

	"github.com/davidebianchi/go-jsonclient"
)

// Options struct define options to create the sdk client
type Options struct {
	APIKey     string
	APICookie  string
	APIBaseURL string
}

// IProjects expose the projects client interface
type IProjects interface {
	Get() (Projects, error)
}

// DeployQuery is used to filter deployment history API.
type DeployQuery struct {
	ProjectID string
}

// IDeploy is a client interface used to interact with deployment pipelines.
type IDeploy interface {
	GetHistory(DeployQuery) ([]DeployItem, error)
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

// New returns the MiaSdkClient to be used to communicate to Mia Platform.
// Console api
func New(opts Options) (*MiaClient, error) {
	if opts.APIKey == "" || opts.APIBaseURL == "" || opts.APICookie == "" {
		return nil, fmt.Errorf("%w: client options are not correct", ErrCreateClient)
	}
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: opts.APIBaseURL,
		Headers: map[string]string{
			"cookie":     opts.APICookie,
			"client-key": opts.APIKey,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCreateClient, err)
	}

	return &MiaClient{
		Projects: &ProjectsClient{
			JSONClient: JSONClient,
		},
	}, nil
}
