package deploy

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	sdkErrors "github.com/mia-platform/miactl/sdk/errors"
)

const (
	SmartDeploy Strategy = "smart_deploy"
	DeployAll   Strategy = "deploy_all"
)

const (
	Created  = "created"
	Pending  = "pending"
	Running  = "running"
	Success  = "success"
	Failed   = "failed"
	Canceled = "canceled"
)

// IDeploy is a client interface used to interact with deployment pipelines.
type IDeploy interface {
	GetHistory(HistoryQuery) ([]Item, error)
	Trigger(string, Config) (Response, error)
	GetDeployStatus(string, int, string) (StatusResponse, error)
}

// Item represents a single item of the deploy history.
type Item struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	Ref         string     `json:"ref"`
	Commit      CommitInfo `json:"commit"`
	User        User       `json:"user"`
	DeployType  string     `json:"deployType"`
	WebURL      string     `json:"webURL"` //nolint:tagliatelle
	Duration    float64    `json:"duration"`
	FinishedAt  time.Time  `json:"finishedAt"`
	Environment string     `json:"env"` //nolint:tagliatelle
}

// CommitInfo represents available information regarding a specific commit.
type CommitInfo struct {
	URL           string    `json:"url"`
	AuthorName    string    `json:"authorName"`
	CommittedDate time.Time `json:"committedDate"`
	AvatarURL     string    `json:"avatarURL"` //nolint:tagliatelle
	Sha           string    `json:"sha"`
}

// User holds the information regarding the user who started
// a specific deploy.
type User struct {
	Name string `json:"name"`
}

// Client implements IDeploy interface to interact with Mia Platform deploy API.
type Client struct {
	JSONClient *jsonclient.Client
}

// Strategy represents the type of deploy strategies that are available on the Console.
type Strategy string

// Config is the details needed to trigger a deploy.
type Config struct {
	Environment         string
	Revision            string
	DeployAll           bool
	ForceDeployNoSemVer bool
}

// Request is the body parameters needed to trigger a pipeline deploy.
type Request struct {
	Environment             string   `json:"environment"`
	Revision                string   `json:"revision"`
	DeployType              Strategy `json:"deployType"`
	ForceDeployWhenNoSemver bool     `json:"forceDeployWhenNoSemver"`
}

// Response is the response of the service after triggering a deploy pipeline.
type Response struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// StatusResponse is the response of the service regarding a deploy pipeline.
type StatusResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// PipelineStatus is one of the possible states in which a deploy pipeline can be found.
type PipelineStatus string

// GetHistory interacts with Mia Platform APIs to retrieve a list of the latest deploy.
func (d Client) GetHistory(query HistoryQuery) ([]Item, error) {
	project, err := getProjectByID(d.JSONClient, query.ProjectID)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("api/deploy/projects/%s/deployment/?page=1&per_page=25&sort=desc", project.ID)

	historyReq, err := d.JSONClient.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var history []Item
	if _, err := d.JSONClient.Do(historyReq, &history); err != nil {
		var httpErr *jsonclient.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", sdkErrors.ErrGeneric, err)
	}
	return history, nil
}

// Trigger interacts with Mia Platform APIs to launch a deploy pipeline with specified configuration.
func (d Client) Trigger(projectID string, cfg Config) (Response, error) {
	data := Request{
		Environment:             cfg.Environment,
		Revision:                cfg.Revision,
		DeployType:              SmartDeploy,
		ForceDeployWhenNoSemver: cfg.ForceDeployNoSemVer,
	}

	if cfg.DeployAll {
		data.DeployType = DeployAll
		data.ForceDeployWhenNoSemver = true
	}

	path := fmt.Sprintf("api/deploy/projects/%s/trigger/pipeline/", projectID)

	request, err := d.JSONClient.NewRequest(http.MethodPost, path, data)
	if err != nil {
		return Response{}, fmt.Errorf("error creating deploy request: %w", err)
	}
	var response Response

	rawRes, err := d.JSONClient.Do(request, &response)
	if err != nil {
		return Response{}, fmt.Errorf("deploy error: %w", err)
	}
	rawRes.Body.Close()

	return response, nil
}

// GetDeployStatus interacts with Mia Platform APIs to retrieve selected pipeline status
func (d Client) GetDeployStatus(projectID string, pipelineID int, environment string) (StatusResponse, error) {
	statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/", projectID, pipelineID)
	if environment != "" {
		statusEndpoint = fmt.Sprintf("%s?environment=%s", statusEndpoint, environment)
	}

	req, err := d.JSONClient.NewRequest(http.MethodGet, statusEndpoint, nil)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("error creating status request: %w", err)
	}

	var statusRes StatusResponse
	rawRes, err := d.JSONClient.Do(req, &statusRes)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("status error: %w", err)
	}
	rawRes.Body.Close()

	return statusRes, nil
}
