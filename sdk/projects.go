package sdk

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
)

// ProjectsClient is the console implementations of the IProjects interface.
type ProjectsClient struct {
	JSONClient *jsonclient.Client
}

// Get method to fetch the console projects.
func (p ProjectsClient) Get() (Projects, error) {
	req, err := p.JSONClient.NewRequest(http.MethodGet, "api/backend/projects/", nil)
	if err != nil {
		return nil, err
	}

	var projects Projects
	var httpErr *jsonclient.HTTPError
	_, err = p.JSONClient.Do(req, &projects)
	if err != nil {
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
	}

	return projects, nil
}
