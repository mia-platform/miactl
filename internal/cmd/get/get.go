package get

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Cluster struct {
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
}

type Environment struct {
	DisplayName string  `json:"label"` //nolint:tagliatelle
	EnvID       string  `json:"value"` //nolint:tagliatelle
	Cluster     Cluster `json:"cluster"`
}
type Pipelines struct {
	Type string `json:"type"`
}

type Project struct {
	ID                   string        `json:"_id"` //nolint:tagliatelle
	Name                 string        `json:"name"`
	ConfigurationGitPath string        `json:"configurationGitPath"`
	Environments         []Environment `json:"environments"`
	ProjectID            string        `json:"projectId"`
	Pipelines            Pipelines     `json:"pipelines"`
	TenantID             string        `json:"tenantId"`
}

const (
	oktaProvider = "okta"
	projectsURI  = "/api/backend/projects/"
)

var (
	validArgs = []string{
		"project", "projects",
		"deployment", "deployments",
	}
	browser = &login.Browser{}
)

// NewGetCmd func creates a new command
func NewGetCmd(options *clioptions.CLIOptions) *cobra.Command {
	return &cobra.Command{
		Use:       "get",
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "projects", "project":
			case "deployment", "deployments":
				return cmd.MarkFlagRequired("project")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := args[0]
			mc := httphandler.NewMiaClientBuilder()
			switch resource {
			case "projects", "project":
				if err := getProjects(mc, options); err != nil {
					return err
				}
			case "deployment", "deployments":
				getDeploymentsForProject()
			}
			return nil
		},
	}
}

func getProjects(mc *httphandler.MiaClient, opts *clioptions.CLIOptions) error {

	// collect base URL from current mia context
	if viper.Get("current-context") == "" {
		return fmt.Errorf("current context is unset")
	}
	currentContext := fmt.Sprint(viper.Get("current-context"))
	baseURL, err := context.GetContextBaseURL(currentContext)
	if err != nil {
		return fmt.Errorf("error retrieving base URL for context %s: %w", currentContext, err)
	}
	// build full path URL
	fullPathURL, err := url.JoinPath(baseURL, projectsURI)
	if err != nil {
		return fmt.Errorf("error building url: %w", err)
	}
	// create a session handler object with the full path URL
	session, err := httphandler.NewSessionHandler(fullPathURL)
	if err != nil {
		return fmt.Errorf("error creating session handler: %w", err)
	}
	// create a new HTTP client and attach it to the session handler
	httpClient, err := httphandler.HTTPClientBuilder(opts)
	if err != nil {
		return fmt.Errorf("error creating HTTP client: %w", err)
	}
	session.WithClient(httpClient)
	// configure authentication
	session.WithAuthentication(baseURL, oktaProvider, browser)
	// attach session handler to mia client
	mc.WithSessionHandler(*session)

	// execute the request
	resp, err := session.Get().ExecuteRequest()
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer resp.Body.Close()

	var projects []Project

	if resp.StatusCode == http.StatusOK {
		companyID, err := context.GetContextCompanyID(currentContext)
		if err != nil {
			return fmt.Errorf("error retrieving company ID for context %s: %w", currentContext, err)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		err = json.Unmarshal(bodyBytes, &projects)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error unmarshaling json response: %w", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Configuration Git Path", "Project ID"})
		for _, project := range projects {
			if project.TenantID == companyID {
				table.Append([]string{project.Name, project.ConfigurationGitPath, project.ProjectID})
			}
		}
		table.Render()
	}

	return nil

}

func getDeploymentsForProject() {

}
