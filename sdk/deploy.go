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

// GetHistory interacts with Mia Platform APIs to retrieve a list of the lastest deploy.
func (d DeployClient) GetHistory(query DeployHistoryQuery) ([]DeployItem, error) {
	project, err := getProjectByID(d.JSONClient, query.ProjectID)
	if err != nil {
		return nil, err
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
