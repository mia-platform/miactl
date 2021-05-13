package deploy

const triggeredPipelinesKey = "triggered-pipelines"

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
