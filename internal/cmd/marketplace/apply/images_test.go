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

package marketplace

import (
	"testing"

	"github.com/mia-platform/miactl/internal/resources/marketplace"
	"github.com/stretchr/testify/require"
)

func TestApplyValidateImageURLs(t *testing.T) {
	t.Run("should throw error with an item that contains both image and imageURL", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]string{
				"localPath": "some/local/path/image.jpg",
			},
			imageURLKey: "http://some.url",
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageURLConflict)
		require.Zero(t, found)
	})

	t.Run("should return local path if element contains image", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]string{
				"localPath": "some/local/path/image.jpg",
			},
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Equal(t, found, "some/local/path/image.jpg")
	})

	t.Run("should return error if image object is not valid", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageKey: map[string]string{
				"someWrongKey": "some/local/path/image.jpg",
			},
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.ErrorIs(t, err, errImageObjectInvalid)
		require.Zero(t, found)
	})
	t.Run("should not return anything if only imageUrl is found", func(t *testing.T) {
		mockItem := &marketplace.Item{
			imageURLKey: "http://some.url",
		}

		found, err := validateAndGetImageLocalPath(mockItem, imageKey, imageURLKey)
		require.NoError(t, err)
		require.Zero(t, found)
	})
}
