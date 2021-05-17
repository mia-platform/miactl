package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
)

const (
	SmartDeploy DeployStrategy = "smart_deploy"
	DeployAll                  = "deploy_all"
)

// DeployItem represents a single item of the deploy history.
type DeployItem struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	Ref         string     `json:"ref"`
	Commit      CommitInfo `json:"commit"`
	User        DeployUser `json:"user"`
	DeployType  string     `json:"deployType"`
	WebURL      string     `json:"webURL"`
	Duration    float64    `json:"duration"`
	FinishedAt  time.Time  `json:"finishedAt"`
	Environment string     `json:"env"`
}

// CommitInfo represents available information regarding a specific commit.
type CommitInfo struct {
	URL        string    `json:"url"`
	AuthorName string    `json:"authorName"`
	CommitDate time.Time `json:"committedDate"`
	AvatarURL  string    `json:"avatarURL"`
	Hash       string    `json:"sha"`
}

// DeployUser holds the information regarding the user who started
// a specific deploy.
type DeployUser struct {
	Name string `json:"name"`
}

// DeployClient implements IDeploy interface to interact with Mia Platform deploy API.
type DeployClient struct {
	JSONClient *jsonclient.Client
}

// DeployStrategy represents the type of deploy strategies that are available on the Console.
type DeployStrategy string

// DeployConfig is the details needed to trigger a deploy.
type DeployConfig struct {
	Environment         string
	Revision            string
	DeployAll           bool
	ForceDeployNoSemVer bool
}

// DeployRequest is the body parameters needed to trigger a pipeline deploy.
type DeployRequest struct {
	Environment             string         `json:"environment"`
	Revision                string         `json:"revision"`
	DeployType              DeployStrategy `json:"deployType"`
	ForceDeployWhenNoSemver bool           `json:"forceDeployWhenNoSemver"`
}

// DeployResponse is the response of the service after triggering a deploy pipeline.
type DeployResponse struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

// GetHistory interacts with Mia Platform APIs to retrieve a list of the latest deploy.
func (d DeployClient) GetHistory(query DeployHistoryQuery) ([]DeployItem, error) {
	project, err := getProjectByID(d.JSONClient, query.ProjectID)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("api/deploy/projects/%s/deployment/?page=1&per_page=25&sort=desc", project.ID)

	historyReq, err := d.JSONClient.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var history []DeployItem
	if _, err := d.JSONClient.Do(historyReq, &history); err != nil {
		var httpErr *jsonclient.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
	}
	return history, nil
}

func (d DeployClient) Trigger(projectId string, cfg DeployConfig) (DeployResponse, error) {
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

	path := fmt.Sprintf("api/deploy/projects/%s/trigger/pipeline/", projectId)

	request, err := d.JSONClient.NewRequest(http.MethodPost, path, data)
	if err != nil {
		return DeployResponse{}, fmt.Errorf("error creating deploy request: %w", err)
	}
	var response DeployResponse

	rawRes, err := d.JSONClient.Do(request, &response)
	if err != nil {
		return DeployResponse{}, fmt.Errorf("deploy error: %w", err)
	}
	rawRes.Body.Close()

	return response, nil
}
