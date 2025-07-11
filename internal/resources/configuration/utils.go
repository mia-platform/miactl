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

package configuration

import (
	"fmt"
	"net/url"
)

func GetEncodedRef(revisionName, versionName string) (string, error) {
	if len(revisionName) > 0 && len(versionName) > 0 {
		return "", fmt.Errorf("both revision and version specified, please provide only one")
	}

	if len(revisionName) > 0 {
		return GetEncodedRevisionRef(revisionName)
	}

	if len(versionName) > 0 {
		return GetEncodedVersionRef(versionName)
	}

	return "", fmt.Errorf("missing revision/version name, please provide one as argument")
}

func GetEncodedRevisionRef(revisionName string) (string, error) {
	if len(revisionName) == 0 {
		return "", fmt.Errorf("missing revision name, please provide a revision name")
	}

	encodedRevisionName := url.PathEscape(revisionName)
	return fmt.Sprintf("revisions/%s", encodedRevisionName), nil
}

func GetEncodedVersionRef(revisionName string) (string, error) {
	if len(revisionName) == 0 {
		return "", fmt.Errorf("missing version name, please provide a version name")
	}

	encodedRevisionName := url.PathEscape(revisionName)
	return fmt.Sprintf("versions/%s", encodedRevisionName), nil
}
