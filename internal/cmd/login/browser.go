package login

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
)

type browserI interface {
	open(string) error
	getEndpoint() string
}

type browser struct {
	endpoint string
}

func (b browser) open(apiUrl string) error {
	if err := open.Run(apiUrl); err != nil {
		fmt.Println("Failed to open browser:", err)
		fmt.Println("Please open the following URL in your browser and complete the authentication process:")
		fmt.Println(apiUrl)
		return err
	}
	return nil
}

func (b browser) getEndpoint() string {
	return b.endpoint
}
