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

package encoding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	JSON = "json"
	YAML = "yaml"

	YamlExtension = ".yaml"
	YmlExtension  = ".yml"
	JSONExtension = ".json"
)

var ErrMarshal = errors.New("error while marshalling data")

type MarshalOptions struct {
	Indent bool
}

type UnsupportedEncodingError struct {
	Encoding string
}

func (e UnsupportedEncodingError) Error() string {
	return fmt.Sprintf("unsupported encoding: %s", e.Encoding)
}

var marshalFuncs = map[string]func(interface{}) ([]byte, error){
	JSON: json.Marshal,
	YAML: yaml.Marshal,
}

func UnmarshalData(data []byte, out interface{}) error {
	return yaml.Unmarshal(data, out)
}

func MarshalData(input interface{}, encoding string, options MarshalOptions) ([]byte, error) {
	marshal, ok := marshalFuncs[encoding]
	if !ok {
		return nil, UnsupportedEncodingError{Encoding: encoding}
	}

	var data []byte
	var err error
	// Intercept panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w: %s", ErrMarshal, r)
			}
		}()
		data, err = marshal(input)
	}()

	if err != nil {
		return nil, err
	}

	if encoding == JSON && options.Indent {
		var indentedData bytes.Buffer
		err := json.Indent(&indentedData, data, "", "  ")
		return indentedData.Bytes(), err
	}

	return data, nil
}
