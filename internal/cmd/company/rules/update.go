package rules

import (
	"fmt"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/files"
	rulesentities "github.com/mia-platform/miactl/internal/resources/rules"
	"github.com/spf13/cobra"
)

func UpdateRules(o *clioptions.CLIOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Company rules",
		Long:  "Update company rules from file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			restConfig, err := o.ToRESTConfig()
			cobra.CheckErr(err)

			client, err := client.APIClientForConfig(restConfig)
			cobra.CheckErr(err)

			rules, err := readFile(o.InputFilePath)
			cobra.CheckErr(err)

			err = New(client).UpdateTenantRules(cmd.Context(), restConfig.CompanyID, rules)
			cobra.CheckErr(err)

			fmt.Printf("Rules updated successfully")
			return nil
		},
	}

	requireFilePathFlag(o, cmd)

	return cmd
}

func requireFilePathFlag(o *clioptions.CLIOptions, cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.InputFilePath, "file-path", "f", "", "paths to JSON file containing the ruleset definition")
	err := cmd.MarkFlagRequired("file-path")
	if err != nil {
		panic(err)
	}
}

func readFile(path string) ([]*rulesentities.SaveChangesRules, error) {
	data := []*rulesentities.SaveChangesRules{}
	if err := files.ReadFile(path, &data); err != nil {
		return data, err
	}

	return data, nil
}
