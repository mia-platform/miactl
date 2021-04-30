package login

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const loginKey = "login"

type tokenRequest struct {
	GrantType  string `json:"grant_type"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	AppID      string `json:"appId"`
	ProviderID string `json:"providerId"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireAt     int64  `json:"expiresAt"`
}

// NewLoginCmd create a new Login command
func NewLoginCmd() *cobra.Command {
	var (
		username   string
		password   string
		appID      string
		providerID string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "authenticate with console",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL := viper.GetString("apibaseurl")
			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}

			accessToken, err := login(baseURL, username, password, appID, providerID)
			if err != nil {
				return err
			}

			// save current token for later commands
			viper.Set("apitoken", accessToken)
			fmt.Println("OK")
			return viper.WriteConfig()
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "your user identifier")
	cmd.Flags().StringVar(&password, "password", "", "your secret password")
	cmd.Flags().StringVar(&appID, "app-id", "", "the type of application is trying to login")
	cmd.Flags().StringVar(&providerID, "provider-id", "", "the authentication provider identifier")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("app-id")
	cmd.MarkFlagRequired("provider-id")

	return cmd
}

func login(authProvider, username, password, appID, providerID string) (string, error) {
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: authProvider,
	})
	if err != nil {
		return "", fmt.Errorf("error creating JSON client: %w", err)
	}

	data := tokenRequest{
		GrantType:  "password",
		Username:   username,
		Password:   password,
		AppID:      appID,
		ProviderID: providerID,
	}
	loginReq, err := JSONClient.NewRequest(http.MethodPost, "/oauth/token", data)
	if err != nil {
		return "", fmt.Errorf("error creating login request: %w", err)
	}
	var loginResponse tokenResponse

	response, err := JSONClient.Do(loginReq, &loginResponse)
	if err != nil || response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth error: %w", err)
	}
	defer response.Body.Close()

	return loginResponse.AccessToken, nil
}
