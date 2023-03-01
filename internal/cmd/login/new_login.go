package login

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	oidc "github.com/coreos/go-oidc"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func NewLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate to the Mia-Platform console",
		RunE: func(cmd *cobra.Command, args []string) error {
			callbackUrl, err := url.Parse("http://localhost:5556/callback")
			if err != nil {
				fmt.Printf("%v", "Error trying to parse the Callback URL")
			}

			server, err := ConfigureAuth("OVERRIDEME", callbackUrl, "OVERRIDEME", "OVERRIDEME")
			if err != nil {
				fmt.Printf("%v", "Error trying to configure OIDC auth")
			}

			server.ListenAndServe()

			return nil
		},
	}
	return cmd
}

func ConfigureAuth(issuerUrl string, callbackUrl *url.URL, clientID string, clientSecret string) (http.Server, error) {
	provider, err := oidc.NewProvider(context.Background(), issuerUrl)
	fmt.Println(err)
	if err != nil {
		return http.Server{}, fmt.Errorf("error creating the provider: %w", err)
	}
	fmt.Println("provider created")

	server := &http.Server{
		Addr: callbackUrl.Host,
	}

	config := oauth2.Config{
		ClientID:     clientID, // change the placeholder with the correct ID of the client
		ClientSecret: clientSecret,
		RedirectURL:  callbackUrl.String(), // change with the correct redirect URL (should be included in the allowed Sign-in redirect URIs list)
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID},
	}

	http.HandleFunc(callbackUrl.Path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		viper.Set("apitoken", token)
		fmt.Println(token)
		if err = viper.WriteConfig(); err != nil {
			fmt.Println("error saving API token in the configuration")
			return
		}

		//closing the server
		err = server.Shutdown(context.Background())
		if err != nil {
			fmt.Printf("%v", err)
		}

	})

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	if err := open.Run(url); err != nil {
		fmt.Println("Failed to open browser:", err)
		fmt.Println("Please open the following URL in your browser and complete the authentication process:")
		fmt.Println(url)
	}

	return *server, nil
}
