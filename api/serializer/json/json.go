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

package json

import (
	"encoding/json"
	"fmt"
	"io"
)

// DefaultEncoding return the encoded object rappresentation in json using the default encoder
func DefaultEncoding(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil && err != io.EOF {
		return []byte{}, fmt.Errorf("error during object encoding: %w", err)
	}

	return data, err
}

// DefaultDecoding parse the encoded body to the given obj using the default decoder
func DefaultDecoding(body []byte, obj interface{}) error {
	err := json.Unmarshal(body, obj)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error during response decoding: %w", err)
	}

	return nil
}
