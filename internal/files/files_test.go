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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSupportedExtension(t *testing.T) {
	testCases := []struct {
		ext      string
		expected bool
	}{
		{ext: ".json", expected: true},
		{ext: ".yml", expected: true},
		{ext: ".yaml", expected: true},
		{ext: ".txt", expected: false},
		{ext: ".pdf", expected: false},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%s => %t", test.ext, test.expected), func(t *testing.T) {
			require.Equal(t, test.expected, isSupportedExtension(test.ext))
		})
	}
}

type TestFileContent struct {
	Answer int `json:"answer"`
}

func TestReadFile(t *testing.T) {
	t.Run("reads json file", func(t *testing.T) {
		content := TestFileContent{}
		require.NoError(t, ReadFile("./testdata/valid.json", &content))
		require.Equal(t, content.Answer, 42)
	})

	t.Run("reads yaml file", func(t *testing.T) {
		content := TestFileContent{}
		require.NoError(t, ReadFile("./testdata/valid.yaml", &content))
		require.Equal(t, content.Answer, 42)
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("file missing", func(t *testing.T) {
			content := TestFileContent{}
			require.ErrorIs(t, ReadFile("./testdata/thisfiledoesnotexist.json", &content), ErrFileReadFailed)
		})

		t.Run("unsupported extension", func(t *testing.T) {
			content := TestFileContent{}
			require.ErrorIs(t, ReadFile("./testdata/unsupported.txt", &content), ErrUnsupportedFile)
		})

		t.Run("malformed file", func(t *testing.T) {
			content := TestFileContent{}
			require.ErrorIs(t, ReadFile("./testdata/malformed.json", &content), ErrFileMalformed)
		})
	})
}
