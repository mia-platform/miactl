package login

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const miactlAppID = "miactl"

type loginConfig struct {
	BaseURL         string
	Username        string
	Password        string
	ProviderID      string
	SkipCertificate bool
}

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
	var cfg loginConfig

	cmd := &cobra.Command{
		Use:   "login",
		Short: "authenticate with console",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cfg.BaseURL = viper.GetString("apibaseurl")
			if cfg.BaseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}

			// set the flag only in case it is defined
			if skipCertificate, err := cmd.Flags().GetBool("insecure"); err == nil {
				cfg.SkipCertificate = skipCertificate
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			accessToken, err := login(cfg)
			if err != nil {
				return err
			}

			// save current token for later commands
			viper.Set("apitoken", accessToken)
			if err = viper.WriteConfig(); err != nil {
				fmt.Println("error saving API token in the configuration")
				return err
			}

			fmt.Println("OK")
			return nil
		},
	}

	cmd.Flags().StringVar(&cfg.Username, "username", "", "your user identifier")
	cmd.Flags().StringVar(&cfg.Password, "password", "", "your secret password")
	cmd.Flags().StringVar(&cfg.ProviderID, "provider-id", "", "the authentication provider identifier")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("provider-id")

	return cmd
}

func login(cfg loginConfig) (string, error) {
	clientOptions := jsonclient.Options{
		BaseURL: cfg.BaseURL,
	}
	if cfg.SkipCertificate {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		clientOptions.HTTPClient = &http.Client{Transport: customTransport}
	}

	JSONClient, err := jsonclient.New(clientOptions)
	if err != nil {
		return "", fmt.Errorf("error creating JSON client: %w", err)
	}

	data := tokenRequest{
		GrantType:  "password",
		Username:   cfg.Username,
		Password:   cfg.Password,
		AppID:      miactlAppID,
		ProviderID: cfg.ProviderID,
	}
	loginReq, err := JSONClient.NewRequest(http.MethodPost, "/api/oauth/token", data)
	if err != nil {
		return "", fmt.Errorf("error creating login request: %w", err)
	}
	var loginResponse tokenResponse

	response, err := JSONClient.Do(loginReq, &loginResponse)
	if err != nil {
		return "", fmt.Errorf("auth error: %w", err)
	}
	defer response.Body.Close()

	return loginResponse.AccessToken, nil
}
