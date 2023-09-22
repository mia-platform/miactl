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
	"errors"

	"github.com/mia-platform/miactl/internal/resources/marketplace"
)

const localPathKey = "localPath"

var (
	errImageURLConflict   = errors.New(`both "image" and "imageUrl" found in the item, only one is admitted`)
	errImageObjectInvalid = errors.New("the image object is not valid")
)

// validateAndGetImageLocalPath looks for an imageKey in the Item, if found it returns the local path
func validateAndGetImageLocalPath(item *marketplace.Item, imageKey, imageURLKey string) (string, error) {
	_, imageURLExists := (*item)[imageURLKey]
	imageInfo, imageExists := (*item)[imageKey]
	if imageExists && imageURLExists {
		return "", errImageURLConflict
	}

	if imageExists {
		localPath, ok := imageInfo.(map[string]string)[localPathKey]
		if !ok {
			return "", errImageObjectInvalid
		}
		return localPath, nil
	}

	return "", nil
}
