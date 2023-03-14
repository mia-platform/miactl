package login

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

type BrowserI interface {
	open(string) error
	getEndpoint() string
}

type Browser struct {
	endpoint string
}

func (b Browser) open(apiUrl string) error {
	if err := open.Run(apiUrl); err != nil {
		fmt.Println("Failed to open browser:", err)
		fmt.Println("Please open the following URL in your browser and complete the authentication process:")
		fmt.Println(apiUrl)
		return err
	}
	return nil
}

func (b Browser) getEndpoint() string {
	return b.endpoint
}
