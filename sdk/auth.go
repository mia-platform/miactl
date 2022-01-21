package sdk

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
)

const appID = "miactl"

// Provider supported by the selected console.
type Provider struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// AuthClient is the console implementations of the IAuth interface
type AuthClient struct {
	JSONClient *jsonclient.Client
}

// GetProviders get the list of oauth providers supported by the target console.
func (c AuthClient) GetProviders(appID string) ([]Provider, error) {
	req, err := c.JSONClient.NewRequest(http.MethodGet, fmt.Sprintf("apps/%s/providers/", appID), nil)
	if err != nil {
		return nil, err
	}

	providers := []Provider{}
	_, err = c.JSONClient.Do(req, &providers)
	if err != nil {
		var httpErr *jsonclient.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
	}

	return providers, nil
}
