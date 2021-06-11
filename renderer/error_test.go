package renderer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/davidebianchi/go-jsonclient"
	sdkErrors "github.com/mia-platform/miactl/sdk/errors"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	t.Run("returns nil if error is nil", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := NewError(buf, nil)
		require.Nil(t, err)
	})

	t.Run("returns error message for generic error", func(t *testing.T) {
		genericErr := fmt.Errorf("%w: test error", sdkErrors.ErrGeneric)
		buf := &bytes.Buffer{}
		errorMessage := NewError(buf, genericErr)
		require.Equal(t, &writeError{
			Message: genericErr.Error(),
			writer:  buf,
		}, errorMessage)
	})

	t.Run("returns error message for create client error", func(t *testing.T) {
		clientErr := fmt.Errorf("%w: test error", sdkErrors.ErrCreateClient)
		buf := &bytes.Buffer{}
		errorMessage := NewError(buf, clientErr)
		require.Equal(t, &writeError{
			Message: clientErr.Error(),
			writer:  buf,
		}, errorMessage)
	})

	t.Run("on http returns error message", func(t *testing.T) {
		httpError := &jsonclient.HTTPError{
			StatusCode: 404,
			Response: &http.Response{
				StatusCode: 404,
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/requested-path",
					},
				},
			},
		}
		buf := &bytes.Buffer{}
		errorMessage := NewError(buf, httpError)
		require.Equal(t, &writeError{
			Message: "GET /requested-path: 404",
			writer:  buf,
		}, errorMessage)
	})

	t.Run("on 401 returns correct message", func(t *testing.T) {
		httpError := &jsonclient.HTTPError{
			StatusCode: 401,
			Response: &http.Response{
				StatusCode: 401,
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/requested-path",
					},
				},
			},
		}
		buf := &bytes.Buffer{}
		errorMessage := NewError(buf, httpError)
		require.Equal(t, &writeError{
			Message: "Unauthorized access, returns 401. Please check your credentials.",
			writer:  buf,
		}, errorMessage)
	})

	t.Run("correctly render message", func(t *testing.T) {
		genericErr := fmt.Errorf("%w: test error", sdkErrors.ErrGeneric)
		buf := &bytes.Buffer{}
		errorMessage := NewError(buf, genericErr)
		errorMessage.Render()

		require.Equal(t, buf.String(), fmt.Sprintf("%s: test error\n", sdkErrors.ErrGeneric))
	})
}

func TestRenderError(t *testing.T) {
	t.Run("if error writer not set, use os.Stdout", func(t *testing.T) {
		err := writeError{
			Message: "my error message",
		}

		out := readFromStdout(err.Render)

		require.Equal(t, fmt.Sprintf("%s\n", err.Message), out)
	})
}

func readFromStdout(funcToCall func()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	funcToCall()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	return out
}
