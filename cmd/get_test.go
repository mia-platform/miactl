package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mia-platform/miactl/factory"
	"github.com/mia-platform/miactl/renderer"
	"github.com/mia-platform/miactl/sdk"
	"github.com/mia-platform/miactl/sdk/deploy"
	sdkErrors "github.com/mia-platform/miactl/sdk/errors"
	"github.com/stretchr/testify/require"
)

var apiKeyValue = "foo"
var apiKeyFlag = fmt.Sprintf(`--apiKey="%s"`, apiKeyValue)
var sidValue = "my-sid"
var apiCookieFlag = fmt.Sprintf(`--apiCookie="sid=%s"`, sidValue)
var apiBaseURLValue = "https://local.io/base-path/"
var apiBaseURLFlag = fmt.Sprintf(`--apiBaseUrl=%s`, apiBaseURLValue)

func TestGetCommandRenderAndReturnsError(t *testing.T) {
	t.Run("without context", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "get", "projects")
		expectedErrMessage := fmt.Sprintf("%s", "context error")
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without correct args", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "get", "not-correct-arg")
		expectedErrMessage := `invalid argument "not-correct-arg" for "miactl get"`
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without args", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "get")
		expectedErrMessage := `accepts 1 arg(s), received 0`
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without required flags", func(t *testing.T) {
		cmd := NewRootCmd()
		ctx := factory.WithValue(context.Background(), cmd.OutOrStdout())
		out, err := executeCommandWithContext(ctx, cmd, "get", "projects")
		expectedErrMessage := fmt.Sprintf("%s: client options are not correct", sdkErrors.ErrCreateClient)
		require.Contains(t, out, expectedErrMessage)
		require.EqualError(t, err, expectedErrMessage)
	})
}

func TestGetCommand(t *testing.T) {
	t.Run("get projects", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{}, "get", "projects", apiKeyFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockProjectsCorrectlyRendered(t, rows)
	})

	t.Run("get project", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{}, "get", "project", apiKeyFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockProjectsCorrectlyRendered(t, rows)
	})

	t.Run("get projects returns error", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{
			ProjectsError: sdkErrors.ErrHTTP,
		}, "get", "projects", apiKeyFlag, apiBaseURLFlag, apiCookieFlag)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("%s\n", sdkErrors.ErrHTTP), out)
	})

}

func TestGetDeployments(t *testing.T) {
	t.Run("returns error if no project ID is provided", func(t *testing.T) {
		out, err := executeRootCommandWithContext(sdk.MockClientError{}, "get", "deployments", apiKeyFlag, apiBaseURLFlag, apiCookieFlag)
		require.Error(t, err)
		require.True(t, strings.HasPrefix(out, "Error: required flag(s) \"project\" not set"))
	})

	var projectIDFlag = fmt.Sprintf("--project=%s", "project-id")
	var projectIDShorthandFlag = fmt.Sprintf("-p=%s", "project-id")

	t.Run("renders error on sdk error", func(t *testing.T) {
		mockErrors := sdk.MockClientError{
			DeployError: fmt.Errorf("Some error"),
		}
		out, err := executeRootCommandWithContext(mockErrors, "get", "deployments", apiKeyFlag, apiBaseURLFlag, apiCookieFlag, projectIDFlag)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(out, "Some error"))
	})

	history := []deploy.DeployItem{
		{
			ID:          123,
			Status:      "running",
			DeployType:  "deploy_all",
			Ref:         "v1.2.3",
			User:        deploy.DeployUser{Name: "John Smith"},
			Duration:    12.3,
			FinishedAt:  time.Date(2020, 01, 12, 22, 33, 44, 12, &time.Location{}),
			WebURL:      "https://web.url/",
			Environment: "development",
		},
		{
			ID:          456,
			Status:      "pending",
			DeployType:  "deploy_all",
			Ref:         "master",
			User:        deploy.DeployUser{Name: "Rick Astley"},
			Duration:    22.99,
			FinishedAt:  time.Date(2020, 02, 12, 22, 33, 44, 12, &time.Location{}),
			WebURL:      "https://web.url.2/",
			Environment: "production",
		},
	}

	t.Run("works with projectId flag", func(t *testing.T) {
		mockErrors := sdk.MockClientError{
			DeployAssertFn: func(query deploy.DeployHistoryQuery) {
				require.Equal(t, deploy.DeployHistoryQuery{
					ProjectID: "project-id",
				}, query)
			},
			DeployHistory: history,
		}
		out, err := executeRootCommandWithContext(mockErrors, "get", "deployments", apiKeyFlag, apiBaseURLFlag, apiCookieFlag, projectIDFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockDeploymentsCorrectlyRendered(t, rows)
	})

	t.Run("works with projectId shorthand flag", func(t *testing.T) {
		mockErrors := sdk.MockClientError{
			DeployAssertFn: func(query deploy.DeployHistoryQuery) {
				require.Equal(t, deploy.DeployHistoryQuery{
					ProjectID: "project-id",
				}, query)
			},
			DeployHistory: history,
		}
		out, err := executeRootCommandWithContext(mockErrors, "get", "deployments", apiKeyFlag, apiBaseURLFlag, apiCookieFlag, projectIDShorthandFlag)
		require.NoError(t, err)
		rows := renderer.CleanTableRows(out)

		assertMockDeploymentsCorrectlyRendered(t, rows)
	})
}

func TestGetProjects(t *testing.T) {
	apiKeyValue := "foo"
	cookieValue := "sid=my-sid"
	apiBaseURLValue := "https://local.io/base-path/"
	mockMiaClient := sdk.WrapperMockMiaClient(sdk.MockClientError{})

	t.Run("render error if get projects throws", func(t *testing.T) {
		buf := &bytes.Buffer{}
		getErr := fmt.Errorf("error getting projects")

		miaClient, err := mockMiaClient(sdk.Options{
			APIKey:     apiKeyValue,
			APICookie:  cookieValue,
			APIBaseURL: apiBaseURLValue,
		})
		require.NoError(t, err)
		prjMock, ok := miaClient.Projects.(*sdk.ProjectsMock)
		require.True(t, ok, "miaClientMock not contains ProjectMock struct")
		prjMock.SetReturnError(getErr)

		f := &factory.Factory{
			Renderer:  renderer.New(buf),
			MiaClient: miaClient,
		}
		getProjects(f)

		require.Equal(t, fmt.Sprintf("%s\n", getErr), buf.String())
	})

	t.Run("render projects table", func(t *testing.T) {
		buf := &bytes.Buffer{}

		miaClient, err := mockMiaClient(sdk.Options{
			APIKey:     apiKeyValue,
			APICookie:  cookieValue,
			APIBaseURL: apiBaseURLValue,
		})
		require.NoError(t, err)

		f := &factory.Factory{
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

func assertMockDeploymentsCorrectlyRendered(t *testing.T, rows []string) {
	expectedHeaders := "# | STATUS | DEPLOY TYPE | ENVIRONMENT | DEPLOY BRANCH/TAG | MADE BY | DURATION | FINISHED AT | VIEW LOG"
	expectedRow1 := "123 | running | deploy_all | development | v1.2.3 | John Smith | 12s | 12 Jan 2020 22:33 UTC | https://web.url/"
	expectedRow2 := "456 | pending | deploy_all | production | master | Rick Astley | 22s | 12 Feb 2020 22:33 UTC | https://web.url.2/"

	require.Lenf(t, rows, 3, "headers + projects")
	require.Equal(t, expectedHeaders, rows[0])
	require.Equal(t, expectedRow1, rows[1])
	require.Equal(t, expectedRow2, rows[2])
}
