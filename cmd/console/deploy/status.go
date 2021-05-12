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

type statusResponse struct {
	PipelineId int    `json:"id"`
	Status     string `json:"status"`
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

			lastEndedDeploy, err := statusMonitor(cmd.OutOrStdout(), baseURL, apiToken, &pipelines)
			if err != nil {
				return err
			}

			viper.Set(triggeredPipelinesKey, pipelines[lastEndedDeploy+1:])
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "all deploy pipelines triggered have completed")
			return nil
		},
	}

	return cmd
}

func statusMonitor(w io.Writer, baseURL, apiToken string, pipelines *sdk.PipelinesConfig) (int, error) {
	lastEndedDeploy := -1
	JSONClient, err := jsonclient.New(jsonclient.Options{
		BaseURL: baseURL,
		Headers: jsonclient.Headers{
			"Authorization": fmt.Sprintf("Bearer %s", apiToken),
		},
	})
	if err != nil {
		return lastEndedDeploy, fmt.Errorf("error creating JSON client: %w", err)
	}

	for i, p := range *pipelines {
		statusEndpoint := fmt.Sprintf("/api/deploy/projects/%s/pipelines/%d/status/?environment=%s", p.ProjectId, p.PipelineId, p.Environment)

		var response statusResponse
		response, err = getStatus(JSONClient, statusEndpoint)
		if err != nil {
			return i, err
		}
		for response.Status != "success" && response.Status != "failed" {
			time.Sleep(2 * time.Second)
			response, err = getStatus(JSONClient, statusEndpoint)
			if err != nil {
				return i, err
			}
		}
		fmt.Fprintf(w, "project: %s\tpipeline: %d\tstatus:%s\n", p.ProjectId, p.PipelineId, response.Status)
		lastEndedDeploy = i
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
