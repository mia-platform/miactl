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

package extensions

import (
	"testing"

	"github.com/mia-platform/miactl/internal/resources/extensibility"
	"github.com/stretchr/testify/require"
)

func TestReadExtensionFromFile(t *testing.T) {
	order := 200.0
	expectedRecord := &extensibility.Extension{
		Name:        "Extension 1",
		Description: "My extension 1",
		Entry:       "https://example.com/",
		Contexts:    []string{"project"},
		Routes: []*extensibility.ExtensionRoute{
			{
				ID:         "extension-1",
				ParentID:   "workloads",
				LocationID: "runtime",
				LabelIntl: map[string]string{
					"en": "SomeLabel",
					"it": "SomeLabelInItalian",
				},
				DestinationPath: "/",
				RenderType:      "menu",
				Order:           &order,
				Icon:            &extensibility.Icon{Name: "PiHardDrives"},
			},
		},
	}

	t.Run("json manifest", func(t *testing.T) {
		ext, err := readExtensionFromFile("./testdata/valid-extension.json")
		require.NoError(t, err)
		require.Equal(t, expectedRecord, ext)
	})

	t.Run("yaml manifest", func(t *testing.T) {
		ext, err := readExtensionFromFile("./testdata/valid-extension.yaml")
		require.NoError(t, err)
		require.Equal(t, expectedRecord, ext)
	})
}
