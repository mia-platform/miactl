package sdk

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
)

const (
	SmartDeploy DeployStrategy = "smart_deploy"
	DeployAll                  = "deploy_all"
)

const (
	Created  PipelineStatus = "created"
	Pending                 = "pending"
	Running                 = "running"
	Success                 = "success"
	Failed                  = "failed"
	Canceled                = "canceled"
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

// StatusResponse is the response of the service regarding a deploy pipeline.
type StatusResponse struct {
	PipelineId int    `json:"id"`
	Status     string `json:"status"`
}

// PipelineStatus is one of the possible states in which a deploy pipeline can be found.
type PipelineStatus string

// Sleeper expose a sleep interface
type Sleeper interface {
	Sleep()
}

// TimeSleeper use time package to implement Sleeper interface
type TimeSleeper struct {
	Delay time.Duration
}

func (ts *TimeSleeper) Sleep() {
	time.Sleep(ts.Delay)
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

// Trigger interacts with Mia Platform APIs to launch a deploy pipeline with specified configuration.
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

// StatusMonitor interacts with Mia Platform APIs to check the status of
// all the pipelines deployed from miactl
func (d DeployClient) StatusMonitor(w io.Writer, pipelines *PipelinesConfig, sl Sleeper) (int, error) {
	lastEndedDeploy := 0

	for _, p := range *pipelines {
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/?environment=%s", p.ProjectId, p.PipelineId, p.Environment)

		var response StatusResponse
		response, err := getStatus(d.JSONClient, statusEndpoint)
		if err != nil {
			return lastEndedDeploy, err
		}
		for shouldRetry(response) {
			sl.Sleep()
			response, err = getStatus(d.JSONClient, statusEndpoint)
			if err != nil {
				return lastEndedDeploy, err
			}
		}
		fmt.Fprintf(w, "project: %s\tpipeline: %d\tstatus:%s\n", p.ProjectId, p.PipelineId, response.Status)
		lastEndedDeploy += 1
	}

	return lastEndedDeploy, nil
}

func getStatus(jc *jsonclient.Client, endpoint string) (StatusResponse, error) {
	req, err := jc.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("error creating status request: %w", err)
	}

	var statusRes StatusResponse

	rawRes, err := jc.Do(req, &statusRes)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("status error: %w", err)
	}
	defer rawRes.Body.Close()

	return statusRes, nil
}

func shouldRetry(sr StatusResponse) bool {
	switch sr.Status {
	case Success:
		return false
	case Failed:
		return false
	case Canceled:
		return false
	default:
		return true
	}
}
