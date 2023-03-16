package get

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/cmd/context"
	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	oktaProvider = "okta"
	projectsURI  = "/api/backend/projects"
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

	session, err := httphandler.NewSessionHandler(projectsURI)
	if err != nil {
		return fmt.Errorf("error creating session handler: %w", err)
	}

	if viper.Get("current-context") == "" {
		return fmt.Errorf("current context is unset")
	}

	currentContext := fmt.Sprint(viper.Get("current-context"))
	baseURL, err := context.GetContextBaseURL(currentContext)
	if err != nil {
		return fmt.Errorf("error retrieving base URL for context %s: %w", currentContext, err)
	}

	session.WithAuthentication(baseURL, oktaProvider, browser)
	mc.WithSessionHandler(*session)

	projects, err := session.Get().ExecuteRequest()
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	fmt.Printf("Projects: %+v\n", projects)
	return nil

	// headers := []string{"#", "Name", "Configuration Git Path", "Project id"}
	// table := f.Renderer.Table(headers)
	// for i, project := range projects {
	// 	table.Append([]string{
	// 		strconv.Itoa(i + 1),
	// 		project.Name,
	// 		project.ConfigurationGitPath,
	// 		project.ProjectID,
	// 	})
	// }
	// table.Render()
}

func getDeploymentsForProject() {

}
