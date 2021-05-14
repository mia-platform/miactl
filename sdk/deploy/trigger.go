package deploy

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/renderer"
	"github.com/spf13/viper"
)

func Trigger(baseUrl, apiToken, projectId string, cfg DeployConfig) (DeployResponse, error) {
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: baseUrl,
		Headers: jsonclient.Headers{
			"Authorization": fmt.Sprintf("Bearer %s", apiToken),
		},
	})
	if err != nil {
		return DeployResponse{}, fmt.Errorf("error creating JSON client: %w", err)
	}

	data := DeployRequest{
		Environment:             cfg.Environment,
		Revision:                cfg.Revision,
		DeployType:              SmartDeploy,
		ForceDeployWhenNoSemver: cfg.ForceDeployNoSemVer,
	}

	if cfg.DeployAll == true {
		data.DeployType = DeployAll
		data.ForceDeployWhenNoSemver = true
	}

	request, err := JSONClient.NewRequest(http.MethodPost, getDeployEndpoint(projectId), data)
	if err != nil {
		return DeployResponse{}, fmt.Errorf("error creating deploy request: %w", err)
	}
	var response DeployResponse

	rawRes, err := JSONClient.Do(request, &response)
	if err != nil {
		return DeployResponse{}, fmt.Errorf("deploy error: %w", err)
	}
	rawRes.Body.Close()

	return response, nil
}

func getDeployEndpoint(projectId string) string {
	return fmt.Sprintf("deploy/projects/%s/trigger/pipeline/", projectId)
}

func VisualizeResponse(r renderer.IRenderer, projectId string, rs DeployResponse) {
	headers := []string{"Project Id", "Deploy Id", "View Pipeline"}
	table := r.Table(headers)
	table.Append([]string{projectId, strconv.FormatInt(int64(rs.Id), 10), rs.Url})
	table.Render()
}

func StorePipelines(ps Pipelines, p Pipeline) error {
	viper.Set(triggeredPipelinesKey, append(ps, p))
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

// ReadPipelines store triggered pipelines details to enabling checking their status
func ReadPipelines(p *Pipelines) error {
	if err := viper.UnmarshalKey(triggeredPipelinesKey, p); err != nil {
		return err
	}

	return nil
}
