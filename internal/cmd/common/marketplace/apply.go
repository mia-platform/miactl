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

import "errors"

var (
	ErrCompanyIDNotDefined = errors.New("companyID must be defined")

	ErrResWithoutName       = errors.New(`the required field "name" was not found in the resource`)
	ErrResWithoutItemID     = errors.New(`the required field "itemId" was not found in the resource`)
	ErrNoValidFilesProvided = errors.New("no valid files were provided")

	ErrResNameNotAString   = errors.New(`the field "name" must be a string`)
	ErrResItemIDNotAString = errors.New(`the field "itemId" must be a string`)

	ErrDuplicatedResIdentifier = errors.New("some resources have duplicated itemId-version tuple")
	ErrUnknownAssetType        = errors.New("unknown asset type")

	ErrUploadingImage    = errors.New("error while uploading image")
	ErrBuildingFilesList = errors.New("error processing files")
	ErrBuildingApplyReq  = errors.New("error preparing apply request")
	ErrProcessingImages  = errors.New("error processing images")
	ErrApplyingResources = errors.New("error applying items")
)

const (
	ImageAssetType = "imageAssetType"
	ImageKey       = "image"
	ImageURLKey    = "imageUrl"

	ItemIDKey = "itemId"

	SupportedByImageAssetType = "supportedByImageAssetType"
	SupportedByImageKey       = "supportedByImage"
	SupportedByImageURLKey    = "supportedByImageUrl"
)
