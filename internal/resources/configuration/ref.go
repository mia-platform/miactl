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
	"errors"
	"fmt"
	"net/url"
)

type RefTypes map[string]bool

const (
	RevisionRefType = "revisions"
	VersionRefType  = "versions"
	BranchRefType   = "branches"
	TagRefType      = "tags"
)

var validRefTypes = RefTypes{RevisionRefType: true, VersionRefType: true, BranchRefType: true, TagRefType: true}

type Ref struct {
	refType string
	refName string
}

func NewRef(refType, refName string) (Ref, error) {
	if !validRefTypes[refType] {
		return Ref{}, fmt.Errorf("unknown reference type: %s", refType)
	}
	if len(refName) == 0 {
		return Ref{}, errors.New("missing reference name, please provide a reference name")
	}
	return Ref{
		refType: refType,
		refName: refName,
	}, nil
}

// EncodedLocationPath returns the encoded path to be used when fetching configuration data
//
// e.g., "<ConsoleURL>/api/projects/<ProjectID>/<EncodedLocationPath()>/configuration"
func (r Ref) EncodedLocationPath() string {
	switch r.refType {
	case RevisionRefType, VersionRefType:
		return fmt.Sprintf("%s/%s", r.refType, url.PathEscape(r.refName))
	case BranchRefType, TagRefType:
		// Legacy projects use /branches endpoint only
		return "branches/" + url.PathEscape(r.refName)
	default:
		return ""
	}
}
