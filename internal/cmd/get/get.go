package get

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
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

// getProjects retrieves the projects with the company ID of the current context
func getProjects(mc *httphandler.MiaClient, opts *clioptions.CLIOptions) error {

	if viper.Get("current-context") == "" {
		return fmt.Errorf("current context is unset")
	}
	currentContext := fmt.Sprint(viper.Get("current-context"))

	session, err := httphandler.ConfigureDefaultSessionHandler(opts, currentContext, projectsURI)
	if err != nil {
		return fmt.Errorf("error building default session handler: %w", err)
	}
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
		if err := httphandler.ParseResponseBody(currentContext, resp.Body, projects); err != nil {
			return fmt.Errorf("error parsing response body: %w", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Configuration Git Path", "Project ID"})
		for _, project := range projects {
			if project.TenantID == companyID {
				table.Append([]string{project.Name, project.ConfigurationGitPath, project.ProjectID})
			}
		}
		table.Render()
	} else {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	return nil

}

func getDeploymentsForProject() {

}
