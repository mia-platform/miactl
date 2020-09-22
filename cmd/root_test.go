package cmd

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/mia-platform/miactl/cmd/factory"
	"github.com/mia-platform/miactl/iostreams"
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

	// TODO: make better
	iostream, _, out, _ := iostreams.Test()
	rootCmd.SetOut(iostream.Out)
	rootCmd.SetErr(iostream.ErrOut)
	rootCmd.SetArgs(args)

	ctx := factory.WithFactoryValueTest(context.Background(), iostream, sdk.WrapperMockMiaClient(mockError))

	err = rootCmd.ExecuteContext(ctx)

	return out.String(), err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
