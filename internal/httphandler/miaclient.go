package httphandler

import (
	"net/http"

	"github.com/mia-platform/miactl/internal/cmd/login"
)

type MiaClient struct {
	request    Request
	httpclient *http.Client
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

func (m *MiaClient) withHttpClient(h *http.Client) *MiaClient {
	m.httpclient = h
	return m
}

func (m *MiaClient) withAuthentication(b login.BrowserI, providerID string, clientUrl string) *MiaClient {
	m.browser = b
	m.providerId = providerID
	m.clientUrl = clientUrl
	return m
}
