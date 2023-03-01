// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package renderer

import (
	"errors"
	"fmt"
	"io"
	"os"

	sdkErrors "github.com/mia-platform/miactl/old/sdk/errors"

	"github.com/davidebianchi/go-jsonclient"
)

const (
	unauthorized = 401
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
	switch {
	case errors.As(err, &httpErr):
		return &writeError{
			Message: httpErrorMessage(httpErr),
			writer:  writer,
		}
	case errors.Is(err, sdkErrors.ErrCreateClient):
		fallthrough
	case errors.Is(err, sdkErrors.ErrGeneric):
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
	case unauthorized:
		return "Unauthorized access, returns 401. Please check your credentials."
	default:
		return httpErr.Error()
	}
}
