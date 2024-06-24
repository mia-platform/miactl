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

package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mia-platform/miactl/internal/encoding"
)

var (
	ErrUnsupportedFile = fmt.Errorf("file extension not supported")
	ErrFileReadFailed  = fmt.Errorf("failed file read")
	ErrFileMalformed   = fmt.Errorf("file malformed")
)

func isSupportedExtension(ext string) bool {
	switch ext {
	case encoding.YmlExtension, encoding.YamlExtension, encoding.JSONExtension:
		return true
	}
	return false
}

func ReadFile(file string, out interface{}) error {
	if !isSupportedExtension(filepath.Ext(file)) {
		return fmt.Errorf("%w: %s", ErrUnsupportedFile, file)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrFileReadFailed, err)
	}

	if err := encoding.UnmarshalData(content, out); err != nil {
		return fmt.Errorf("%w %s: %s", ErrFileMalformed, file, err)
	}
	return nil
}
