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

package errors

import (
	"errors"

	"github.com/davidebianchi/go-jsonclient"
)

var (
	// ErrGeneric is a generic error
	ErrGeneric = errors.New("Something went wrong")
	// ErrHTTP is an http error
	ErrHTTP = jsonclient.ErrHTTP
	// ErrCreateClient is the error creating MiaSdkClient
	ErrCreateClient = errors.New("Error creating sdk client")
	// ErrProjectNotFound is the error returned when specified
	// project cannot be found.
	ErrProjectNotFound = errors.New("Project not found")
)
