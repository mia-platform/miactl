package login

import (
	"fmt"
	"net/http"
)

type fakeBrowser struct {
	code        string
	state       string
	callbackUrl string
	endpoint    string
}

func (f fakeBrowser) open(apiUrl string) error {
	http.DefaultClient.Get(fmt.Sprintf("http://%s/oauth/callback?code=%s&state=%s", f.callbackUrl, f.code, f.state))

	return nil
}

func (b fakeBrowser) getEndpoint() string {
	return b.endpoint
}
