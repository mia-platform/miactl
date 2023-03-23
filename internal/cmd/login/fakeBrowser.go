package login

import (
	"fmt"
	"net/http"
)

type FakeBrowser struct {
	code        string
	state       string
	callbackUrl string
	endpoint    string
}

func (f FakeBrowser) open(apiUrl string) error {
	go func() {
		http.DefaultClient.Get(fmt.Sprintf("http://%s/oauth/callback?code=%s&state=%s", f.callbackUrl, f.code, f.state))
	}()
	return nil
}

func (b FakeBrowser) getEndpoint() string {
	return b.endpoint
}
