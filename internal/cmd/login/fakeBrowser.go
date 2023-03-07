package login

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type fakeBrowser struct {
	code        string
	state       string
	callbackUrl string
	endpoint    string
}

func (f fakeBrowser) open(apiUrl string) error {
	r, err := http.DefaultClient.Get(fmt.Sprintf("%s/oauth/callback?code=%s&state=%s", f.callbackUrl, f.code, f.state))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	fmt.Println(body)
	return nil

}

func (b fakeBrowser) getEndpoint() string {
	return b.endpoint
}
