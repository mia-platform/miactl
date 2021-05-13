package deploy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/davidebianchi/go-jsonclient"
	"github.com/mia-platform/miactl/sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const checkDelay = 2 * time.Second
const endMessage = "all deploy pipelines triggered have completed"

const (
	Created  PipelineStatus = "created"
	Pending                 = "pending"
	Running                 = "running"
	Success                 = "success"
	Failed                  = "failed"
	Canceled                = "canceled"
)

type PipelineStatus string

type statusResponse struct {
	PipelineId int    `json:"id"`
	Status     string `json:"status"`
}

type sleeper interface {
	Sleep()
}

type timeSleeper struct {
	delay time.Duration
}

func (ts *timeSleeper) Sleep() {
	time.Sleep(ts.delay)
}

func NewStatusCmd() *cobra.Command {
	var (
		baseURL  string
		apiToken string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "verify status of deploy pipeline",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			baseURL = viper.GetString("apibaseurl")
			apiToken = viper.GetString("apitoken")

			if apiToken == "" {
				return errors.New("missing API token - please login")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				return errors.New("API base URL not specified nor configured")
			}

			var pipelines sdk.PipelinesConfig
			if err := viper.UnmarshalKey(triggeredPipelinesKey, &pipelines); err != nil {
				return err
			}

			if len(pipelines) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no deploy pipelines triggered found")
				return nil
			}

			tSleep := &timeSleeper{checkDelay}
			lastEndedDeploy, err := statusMonitor(cmd.OutOrStdout(), baseURL, apiToken, &pipelines, tSleep)
			if err != nil {
				return err
			}

			viper.Set(triggeredPipelinesKey, pipelines[lastEndedDeploy:])
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "all deploy pipelines triggered have completed")
			return nil
		},
	}

	return cmd
}

func statusMonitor(w io.Writer, baseURL, apiToken string, pipelines *sdk.PipelinesConfig, sl sleeper) (int, error) {
	lastEndedDeploy := 0
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: baseURL,
		Headers: jsonclient.Headers{
			"Authorization": fmt.Sprintf("Bearer %s", apiToken),
		},
	})
	if err != nil {
		return lastEndedDeploy, fmt.Errorf("error creating JSON client: %w", err)
	}

	for _, p := range *pipelines {
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/?environment=%s", p.ProjectId, p.PipelineId, p.Environment)

		var response statusResponse
		response, err = getStatus(JSONClient, statusEndpoint)
		if err != nil {
			return lastEndedDeploy, err
		}
		for shouldRetry(response) {
			sl.Sleep()
			response, err = getStatus(JSONClient, statusEndpoint)
			if err != nil {
				return lastEndedDeploy, err
			}
		}
		fmt.Fprintf(w, "project: %s\tpipeline: %d\tstatus:%s\n", p.ProjectId, p.PipelineId, response.Status)
		lastEndedDeploy += 1
	}

	return lastEndedDeploy, nil
}

func getStatus(jc *jsonclient.Client, endpoint string) (statusResponse, error) {
	req, err := jc.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return statusResponse{}, fmt.Errorf("error creating status request: %w", err)
	}

	var statusRes statusResponse

	rawRes, err := jc.Do(req, &statusRes)
	if err != nil {
		return statusResponse{}, fmt.Errorf("status error: %w", err)
	}
	defer rawRes.Body.Close()

	return statusRes, nil
}

func shouldRetry(sr statusResponse) bool {
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
