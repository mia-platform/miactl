package login

import (
	"fmt"
	"net/http"

	"github.com/mia-platform/miactl/browser"
	"github.com/mia-platform/miactl/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/davidebianchi/go-jsonclient"
	"github.com/spf13/cobra"
)

const (
	appID = "miactl"

	hostedConsoleResponse = "hosted console"
	paasConsole           = "console.cloud.mia-platform.eu"
)

type Provider struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// NewLoginCmd returns the login command
func NewLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with console",
		RunE: func(cmd *cobra.Command, args []string) error {
			// f, err := factory.GetFactoryFromContext(cmd.Context(), opts)
			// if err != nil {
			// 	return err
			// }

			consoleHost := ""
			err := prompt.AskOneSurvey(&survey.Select{
				Message: "Select a console:",
				Options: []string{paasConsole, hostedConsoleResponse},
			}, &consoleHost)
			if err != nil {
				return err
			}

			if consoleHost == hostedConsoleResponse {
				err := prompt.AskOneSurvey(&survey.Input{
					Message: "Write console hostname",
					Help:    "Write it without protocol",
				}, &consoleHost)
				if err != nil {
					return err
				}
			}

			consoleHost = fmt.Sprintf("https://%s", consoleHost)

			jsonClient, err := jsonclient.New(jsonclient.Options{BaseURL: fmt.Sprintf("%s/api/", consoleHost)})
			if err != nil {
				return err
			}

			req, err := jsonClient.NewRequest(http.MethodGet, fmt.Sprintf("apps/%s/providers/", appID), nil)
			if err != nil {
				return err
			}

			providers := []Provider{}
			_, err = jsonClient.Do(req, &providers)
			if err != nil {
				return err
			}
			if len(providers) == 0 {
				return fmt.Errorf("not exists auth providers for app %s", appID)
			}

			providerID := providers[0].ID
			if len(providers) > 1 {
				var providersIDs []string
				for _, provider := range providers {
					providersIDs = append(providersIDs, provider.ID)
				}

				err = prompt.AskOneSurvey(&survey.Select{
					Message: "Select the provider to authenticate with:",
					Options: providersIDs,
					Default: providersIDs[0],
				}, &providerID)
				if err != nil {
					return err
				}
			}

			return loginRun(jsonClient, consoleHost, providerID)
		},
	}
}

func loginRun(jsonClient *jsonclient.Client, consoleHost, providerID string) error {
	o := oauth{
		OpenBrowser: func(url string) error {
			cmd, err := browser.Open(url)
			if err != nil {
				return err
			}
			return cmd.Run()
		},
		HTTPClient: jsonClient,

		localServerAddress: "127.0.0.1:53535",
	}

	token, err := o.localServerFlow(consoleHost, appID, providerID)
	if err != nil {
		return err
	}

	fmt.Printf("ACCESS TOKEN: %s", token.AccessToken)

	return nil
}
