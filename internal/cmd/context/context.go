package context

import (
	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/spf13/cobra"
)

type contextFlags struct {
	RootOptions *clioptions.RootOptions
}

type contextOptions struct {
	name      string
	endpoint  string
	projectID string
	companyID string
}

func NewContextCmd() *cobra.Command {
	flags := newContextFlags()
	cmd := &cobra.Command{
		Use:   "context",
		Short: "perform operations on cluster contexts",
	}

	cmd.AddCommand(NewSetContextCmd())
	cmd.AddCommand(NewUseContextCmd())

	flags.addFlags(cmd)

	return cmd
}

func (f *contextFlags) addFlags(c *cobra.Command) {
	//root flags
	f.RootOptions.AddFlags(c)
}

func newContextFlags() *contextFlags {
	return &contextFlags{
		RootOptions: clioptions.NewRootOptions(),
	}
}

func getOptions(c *cobra.Command) *contextOptions {
	return &contextOptions{
		name:      clioptions.GetFlagString(c, "context"),
		endpoint:  clioptions.GetFlagString(c, "endpoint"),
		projectID: clioptions.GetFlagString(c, "project-id"),
		companyID: clioptions.GetFlagString(c, "company-id"),
	}
}
