package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"

	"github.com/stretchr/testify/require"
)

func TestGetCommandRenderAndReturnsError(t *testing.T) {
	t.Run("without context", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "get", "projects")
		expectedErrMessage := fmt.Sprintf("%s", "context error")
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without correct args", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "get", "not-correct-arg")
		expectedErrMessage := `invalid argument "not-correct-arg" for "miaplatformctl get"`
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without required flags", func(t *testing.T) {
		cmd := NewRootCmd()
		ctx := WithFactoryValue(context.Background(), cmd.OutOrStdout())
		out, err := executeCommandWithContext(ctx, cmd, "get", "projects")
		expectedErrMessage := fmt.Sprintf("%s: client options are not correct", sdk.ErrCreateClient)
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})
}

func TestGetCommand(t *testing.T) {
	secretValue := "foo"
	secretFlag := fmt.Sprintf(`--apiKey="%s"`, secretValue)
	sidValue := "my-sid"
	apiCookieFlag := fmt.Sprintf(`--apiCookie="sid=%s"`, sidValue)
	apiBaseURLValue := "https://local.io/base-path/"
	apiBaseURLFlag := fmt.Sprintf(`--apiBaseUrl=%s`, apiBaseURLValue)

	t.Run("get projects", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{}, "get", "projects", secretFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockProjectsCorrectlyRendered(t, rows)
	})

	t.Run("get project", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{}, "get", "project", secretFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockProjectsCorrectlyRendered(t, rows)
	})

	t.Run("get projects returns error", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{
			ProjectsError: sdk.ErrHTTP,
		}, "get", "projects", secretFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("%s\n", sdk.ErrHTTP), out)
	})
}

func TestGetProjects(t *testing.T) {
	secretValue := "foo"
	cookieValue := "sid=my-sid"
	apiBaseURLValue := "https://local.io/base-path/"
	mockMiaClient := sdk.WrapperMockMiaClient(sdk.MockClientError{})

	t.Run("render error if get projects throws", func(t *testing.T) {
		buf := &bytes.Buffer{}
		getErr := fmt.Errorf("error getting projects")

		miaClient, err := mockMiaClient(sdk.Options{
			APIKey:     secretValue,
			APICookie:  cookieValue,
			APIBaseURL: apiBaseURLValue,
		})
		require.NoError(t, err)
		prjMock, ok := miaClient.Projects.(*sdk.ProjectsMock)
		require.True(t, ok, "miaClientMock not contains ProjectMock struct")
		prjMock.SetReturnError(getErr)

		f := &Factory{
			Renderer:  renderer.New(buf),
			MiaClient: miaClient,
		}
		getProjects(f)

		require.Equal(t, fmt.Sprintf("%s\n", getErr), buf.String())
	})

	t.Run("render projects table", func(t *testing.T) {
		buf := &bytes.Buffer{}

		miaClient, err := mockMiaClient(sdk.Options{
			APIKey:     secretValue,
			APICookie:  cookieValue,
			APIBaseURL: apiBaseURLValue,
		})
		require.NoError(t, err)

		f := &Factory{
			Renderer:  renderer.New(buf),
			MiaClient: miaClient,
		}
		getProjects(f)

		rows := renderer.CleanTableRows(buf.String())
		assertMockProjectsCorrectlyRendered(t, rows)
	})
}

func assertMockProjectsCorrectlyRendered(t *testing.T, rows []string) {
	projectsMock := sdk.ProjectsMock{}
	projects, err := projectsMock.Get()
	require.NoError(t, err)

	require.Lenf(t, rows, 1+len(projects), "headers + projects")

	expectedHeaders := "# | NAME | CONFIGURATION GIT PATH | PROJECT ID"
	expectedRow1 := "1 | Project 1 | /git/path | project-1"
	expectedRow2 := "2 | Project 2 | /git/path | project-2"

	require.Equal(t, expectedHeaders, rows[0])
	require.Equal(t, expectedRow1, rows[1])
	require.Equal(t, expectedRow2, rows[2])
}
