package login

import (
	"fmt"

	"github.com/mia-platform/miactl/browser"
	"github.com/mia-platform/miactl/sdk"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/spf13/cobra"
)

// NewLoginCmd returns the login command
func NewLoginCmd(opts sdk.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with console",
		RunE: func(cmd *cobra.Command, args []string) error {
			// f, err := factory.GetFactoryFromContext(cmd.Context(), opts)
			// if err != nil {
			// 	return err
			// }

			return loginRun()
		},
	}
}

func loginRun() error {
	hostname := "https://console.test.mia-platform.eu"
	jsonClient, err := jsonclient.New(jsonclient.Options{})
	if err != nil {
		return err
	}

	oauth := Oauth{
		OpenBrowser: browser.OpenBrowser,
		HTTPClient:  jsonClient,
	}

	accessToken, err := oauth.localServerFlow(hostname, "miactl", "gitlab")
	if err != nil {
		return err
	}

	fmt.Printf("ACCESS TOKEN: %s", accessToken)

	return nil
}
