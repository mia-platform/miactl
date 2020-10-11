package sdk

import (
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

type AuthClient struct {
	JSONClient *jsonclient.Client
}

func (c AuthClient) GetProviders(appID string) ([]Provider, error) {
	req, err := c.JSONClient.NewRequest(http.MethodGet, fmt.Sprintf("apps/%s/providers/", appID), nil)
	if err != nil {
		return nil, err
	}

	providers := []Provider{}
	_, err = c.JSONClient.Do(req, &providers)
	if err != nil {
		return nil, err
	}
	return providers, nil
}
