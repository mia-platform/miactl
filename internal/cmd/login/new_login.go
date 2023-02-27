package login

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
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
			provider, err := oidc.NewProvider(context.Background(), "https://login.mia-platform.eu/oauth2/aus2m8z66r5k6Cbip417")
			if err != nil {
				return fmt.Errorf("error creating the provider: %w", err)
			}

			config := oauth2.Config{
				ClientID:     "client-id", // change the placeholder with the correct ID of the client
				ClientSecret: "client-secret",
				RedirectURL:  "https://redirect-url", // change with the correct redirect URL (should be included in the allowed Sign-in redirect URIs list)
				Endpoint:     provider.Endpoint(),
			}

			http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
				code := r.URL.Query().Get("code")
				token, err := config.Exchange(context.Background(), code)
				if err != nil {
					http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
					return
				}
				viper.Set("apitoken", token)
				if err = viper.WriteConfig(); err != nil {
					fmt.Println("error saving API token in the configuration")
					return
				}
			})

			url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
			if err := open.Run(url); err != nil {
				fmt.Println("Failed to open browser:", err)
				fmt.Println("Please open the following URL in your browser and complete the authentication process:")
				fmt.Println(url)
			}

			return nil
		},
	}
	return cmd
}
