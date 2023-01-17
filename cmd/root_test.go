package cmd

import (
	"bytes"
	"context"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/sdk"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandWithContext(ctx context.Context, root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)

	return buf.String(), err
}

func executeRootCommandWithContext(mockError sdk.MockClientError, args ...string) (output string, err error) {
	rootCmd := NewRootCmd()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	ctx := factory.WithValueTest(
		context.Background(),
		rootCmd.OutOrStderr(),
		sdk.WrapperMockMiaClient(mockError),
	)

	err = rootCmd.ExecuteContext(ctx)

	return buf.String(), err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
