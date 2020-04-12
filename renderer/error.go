package renderer

import (
	"errors"
	"fmt"
	"io"
	"os"

	"miactl/sdk"

	"github.com/davidebianchi/go-jsonclient"
)

// IError is the error interface
type IError interface {
	Render()
}

type writeError struct {
	Message string

	writer io.Writer
}

// Render method should be called to display the correct error message
func (e *writeError) Render() {
	if e.writer == nil {
		e.writer = os.Stdout
	}
	fmt.Fprintln(e.writer, e.Message)
}

// NewError returns the error with the correct message
func NewError(writer io.Writer, err error) IError {
	if err == nil {
		return nil
	}
	var httpErr *jsonclient.HTTPError
	switch true {
	case errors.As(err, &httpErr):
		return &writeError{
			Message: httpErrorMessage(httpErr),
			writer:  writer,
		}
	case errors.Is(err, sdk.ErrCreateClient):
		fallthrough
	case errors.Is(err, sdk.ErrGeneric):
		fallthrough
	default:
		return &writeError{
			Message: err.Error(),
			writer:  writer,
		}
	}
}

func httpErrorMessage(httpErr *jsonclient.HTTPError) string {
	switch httpErr.StatusCode {
	case 401:
		return "Unauthorized access, returns 401. Please check your credentials."
	default:
		return httpErr.Error()
	}
}
