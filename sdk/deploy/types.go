package deploy

type DeployStrategy string

const (
	SmartDeploy DeployStrategy = "smart_deploy"
	DeployAll                  = "deploy_all"
)

const triggeredPipelinesKey = "triggered-pipelines"

type DeployConfig struct {
	Environment         string
	Revision            string
	DeployAll           bool
	ForceDeployNoSemVer bool
}

type DeployRequest struct {
	Environment             string         `json:"environment"`
	Revision                string         `json:"revision"`
	DeployType              DeployStrategy `json:"deployType"`
	ForceDeployWhenNoSemver bool           `json:"forceDeployWhenNoSemver"`
}

type DeployResponse struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

type Pipeline struct {
	ProjectId   string `yaml:"projectid" mapstructure:"projectid"`
	PipelineId  int    `yaml:"pipelineid" mapstructure:"pipelineid"`
	Environment string `yaml:"environment" mapstructure:"environment"`
}

type Pipelines []Pipeline

type statusResponse struct {
	PipelineId int    `json:"id"`
	Status     string `json:"status"`
}
