package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var validArgs = []string{
	"project", "projects",
	"deploy", "deploys",
}

// NewGetCmd func creates a new command
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "get",
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactValidArgs(1)(cmd, args)
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
			case "deploy", "deploys":
				getDeploysForProject(cmd)
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

func getDeploysForProject(cmd *cobra.Command) {
	// cmd.PersistentFlags().StringVar()

	fmt.Printf("here we will use the DeployClient from factory\n")

}
