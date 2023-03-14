package httphandler

import (
	"github.com/mia-platform/miactl/internal/cmd/login"
)

type MiaClient struct {
	request    Request
	browser    login.BrowserI
	providerId string
	clientUrl  string
}

func NewMiaClientBuilder() *MiaClient {
	return &MiaClient{}
}

func (m *MiaClient) withRequest(r Request) *MiaClient {
	m.request = r
	return m
}

func (m *MiaClient) withAuthentication(b login.BrowserI, providerID string, clientUrl string) *MiaClient {
	m.browser = b
	m.providerId = providerID
	m.clientUrl = clientUrl
	return m
}
