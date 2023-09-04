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
	"reflect"
	"testing"
)

type TestStruct struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

type FaultyTestStruct struct {
	Name      string `json:"name" yaml:"name"`
	Value     int    `json:"value" yaml:"value"`
	NotSigned func()
}

func TestUnsupportedEncodingError(t *testing.T) {
	err := UnsupportedEncodingError{Encoding: "txt"}
	want := "unsupported encoding: txt"

	if got := err.Error(); got != want {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestUnmarshalData(t *testing.T) {
	tests := []struct {
		data     []byte
		encoding string
		want     TestStruct
		wantErr  bool
	}{
		{[]byte(`{"name": "Alice", "value": 42}`), JSON, TestStruct{"Alice", 42}, false},
		{[]byte("name: Alice\nvalue: 42"), YAML, TestStruct{"Alice", 42}, false},
		{[]byte(`{"name": "Alice", "value": 42}`), "txt", TestStruct{}, true},
	}

	for _, test := range tests {
		var got TestStruct
		err := UnmarshalData(test.data, test.encoding, &got)

		if (err != nil) != test.wantErr {
			t.Errorf("UnmarshalData(%v, %v) returned error %v; wantErr %v", test.data, test.encoding, err, test.wantErr)
			continue
		}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("UnmarshalData(%v, %v) = %v; want %v", test.data, test.encoding, got, test.want)
		}
	}
}

func TestMarshalData(t *testing.T) {
	data := TestStruct{"Bob", 24}
	faultyData := FaultyTestStruct{"Bob", 24, func() {}}
	tests := []struct {
		input    interface{}
		encoding string
		opts     MarshalOptions
		wantErr  bool
	}{
		{data, JSON, MarshalOptions{}, false},
		{data, YAML, MarshalOptions{}, false},
		{data, "txt", MarshalOptions{}, true},
		{data, JSON, MarshalOptions{Indent: true}, false},
		{faultyData, JSON, MarshalOptions{}, true},
		{faultyData, YAML, MarshalOptions{}, true},
	}

	for _, test := range tests {
		got, err := MarshalData(test.input, test.encoding, test.opts)

		if (err != nil) != test.wantErr {
			t.Errorf("MarshalData(%v, %v) returned error %v; wantErr %v", test.input, test.encoding, err, test.wantErr)
			continue
		}

		if err == nil && len(got) == 0 {
			t.Errorf("MarshalData(%v, %v) returned empty byte slice", test.input, test.encoding)
		}
	}
}
