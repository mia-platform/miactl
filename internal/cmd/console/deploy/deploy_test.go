package deploy

import (
	"fmt"
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/mia-platform/miactl/internal/httphandler"
	"github.com/mia-platform/miactl/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewDeployCmd(t *testing.T) {
	t.Run("test command creation", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		cmd := NewDeployCmd(opts)
		require.NotNil(t, cmd)
	})
}

func TestInitializeClient(t *testing.T) {
	viper.New()
	viper.Set("contexts.ctx.apibaseurl", "apibaseurl")
	viper.Set("contexts.ctx.projectid", "projectid")
	viper.Set("contexts.ctx.companyid", "companyid")

	t.Run("Inizialize a client succesfully", func(t *testing.T) {
		opts := clioptions.NewCLIOptions()
		ep := "endpoint"
		ctx := "ctx"

		mc, err := initializeClient(opts, ep, ctx)

		require.NoError(t, err)
		require.NotNil(t, mc)

	})

}

func TestTriggerPipeline(t *testing.T) {
	server := testutils.CreateMockServer()
	server.Start()

	defer server.Close()

	opts := clioptions.CLIOptions{
		Revision:            "test",
		DeployType:          "smart-deploy",
		ForceDeployNoSemVer: true,
	}
	mc := httphandler.FakeMiaClient(fmt.Sprintf("%s/api/deploy/projects/projectid/trigger/pipeline/", server.URL))
	t.Run("Trigger succesfully a pipeline", func(t *testing.T) {
		exectedBody := deployRespnse{
			Id:  123,
			Url: "pipeline.eu",
		}

		body, err := triggerPipeline(mc, "fake-env", &opts)
		if err != nil {
			fmt.Println(err)
		}
		require.Equal(t, *body, exectedBody)
	})

	t.Run("Trigger succesfully a pipeline with deploy_all", func(t *testing.T) {
		opts.DeployType = "deploy_all"
		exectedBody := deployRespnse{
			Id:  123,
			Url: "pipeline.eu",
		}

		body, err := triggerPipeline(mc, "fake-env", &opts)
		if err != nil {
			fmt.Println(err)
		}
		require.Equal(t, *body, exectedBody)
	})
	mc = httphandler.FakeMiaClient(fmt.Sprintf("%s/notfound", server.URL))
	t.Run("Trigger a pipeline with response status 404", func(t *testing.T) {
		_, err := triggerPipeline(mc, "fake-env", &opts)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(err)
		require.ErrorContains(t, err, "pipeline exited with status: 404")
	})
}

func TestWaitStatus(t *testing.T) {
	server := testutils.CreateMockServer()
	server.Start()

	defer server.Close()
	t.Run("wait succesfully for pipeline completion", func(t *testing.T) {
		mc := httphandler.FakeMiaClient(fmt.Sprintf("%s/api/deploy/projects/projectid/pipelines/123/status/", server.URL))

		result, err := waitStatus(mc)
		if err != nil {
			fmt.Println(err)
		}
		require.Equal(t, "succeed", result)

	})
}

func TestRun(t *testing.T) {

	opts := clioptions.NewCLIOptions()
	opts.ProjectID = "projectid"
	t.Run("run succesfully", func(t *testing.T) {
		err := run("fake-env", opts, initMiaClientWithURL)
		if err != nil {
			panic(err)
		}

		require.NoError(t, err)
	})
}

func initMiaClientWithURL(opts *clioptions.CLIOptions, endpoint string, currentContext string) (*httphandler.MiaClient, error) {
	server := testutils.CreateMockServer()
	server.Start()

	url := fmt.Sprintf("%s%s", server.URL, endpoint)
	mc := httphandler.FakeMiaClient(url)

	return mc, nil
}
