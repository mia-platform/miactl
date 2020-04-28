package sdk

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
)

// DeployItem represents a single item of the deploy history.
type DeployItem struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	Ref         string     `json:"ref"`
	Commit      CommitInfo `json:"commit"`
	User        DeployUser `json:"iser"`
	DeployType  string     `json:"deployType"`
	WebURL      string     `json:"webURL"`
	Duration    float64    `json:"duration"`
	FinishedAt  time.Time  `json:"finishedAt"`
	Environment string     `json:"environment"`
}

// CommitInfo represents available information regarding a specific commit.
type CommitInfo struct {
	URL        string    `json:"url"`
	AuthorName string    `json:"authorName"`
	CommitDate time.Time `json:"commitDate"`
	AvatarURL  string    `json:"avatarURL"`
	Hash       string    `json:"sha"`
}

// DeployUser holds the information regarding the user who started
// a specific deploy.
type DeployUser struct {
	Name string `json:"name"`
}

type DeployClient struct {
	JSONClient *jsonclient.Client
}

func (d DeployClient) GetHistory(projectID string) ([]DeployItem, error) {
	req, err := d.JSONClient.NewRequest(http.MethodGet, "api/backend/projects/", nil)
	if err != nil {
		return nil, err
	}

	var projects Projects
	if _, err := d.JSONClient.Do(req, &projects); err != nil {
		var httpErr *jsonclient.HTTPError
		if errors.As(err, &httpErr) {
			return nil, httpErr
		}
		return nil, fmt.Errorf("%w: %s", ErrGeneric, err)
	}

	var project *Project
	for _, p := range projects {
		if p.ProjectID == projectID {
			project = &p
			break
		}
	}
	if project == nil {
		return nil, fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
	}

	path := fmt.Sprintf("api/backend/projects/%s/deployment/?page=1&per_page=25&sort=desc", project.ID)

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
