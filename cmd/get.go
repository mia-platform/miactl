package cmd

import (
	"strconv"
	"time"

	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/cobra"
)

var validArgs = []string{
	"project", "projects",
	"deployment", "deployments",
}

// NewGetCmd func creates a new command
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "get",
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "projects", "project":
			case "deployment", "deployments":
				cmd.MarkFlagRequired("project")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := GetFactoryFromContext(cmd.Context(), opts)
			if err != nil {
				return err
			}

			resource := args[0]

			switch resource {
			case "projects", "project":
				getProjects(f)
			case "deployment", "deployments":
				getDeploysForProject(f)
			}
			return nil
		},
	}
}

func getProjects(f *Factory) {
	projects, err := f.MiaClient.Projects.Get()
	if err != nil {
		f.Renderer.Error(err).Render()
		return
	}

	headers := []string{"#", "Name", "Configuration Git Path", "Project id"}
	table := f.Renderer.Table(headers)
	for i, project := range projects {
		table.Append([]string{
			strconv.Itoa(i + 1),
			project.Name,
			project.ConfigurationGitPath,
			project.ProjectID,
		})
	}
	table.Render()
}

func getDeploysForProject(f *Factory) {
	query := sdk.DeployHistoryQuery{
		ProjectID: projectID,
	}

	history, err := f.MiaClient.Deploy.GetHistory(query)
	if err != nil {
		f.Renderer.Error(err).Render()
		return
	}

	headers := []string{"#", "Status", "Deploy Type", "Deploy Branch/Tag", "Made By", "Duration", "Finished At", "View Log"}
	table := f.Renderer.Table(headers)
	for _, deploy := range history {
		table.Append([]string{
			strconv.Itoa(deploy.ID),
			deploy.Status,
			deploy.DeployType,
			deploy.Ref,
			deploy.User.Name,
			time.Duration(time.Duration(deploy.Duration) * time.Second).String(),
			renderer.FormatDate(deploy.FinishedAt),
			deploy.WebURL,
		})
	}
	table.Render()
}
