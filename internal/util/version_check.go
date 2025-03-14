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

package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	defaultMajor = "13"
	defaultMinor = "6"
)

var cachedVersions = map[string]*resources.Version{}

// VersionCheck will check if the remote version endpoint is greater or not of major and minor version passed
func VersionCheck(ctx context.Context, client client.Interface, major, minor int) (bool, error) {
	remoteMajor, remoteMinor, err := remoteVersion(ctx, client)
	if err != nil {
		return false, err
	}

	return versionCompare(remoteMajor, remoteMinor, major, minor), nil
}

func remoteVersion(ctx context.Context, client client.Interface) (string, string, error) {
	request := client.Get().APIPath("/api/version")
	cacheKey := request.URL().String()
	if version, found := getFromCache(request.URL().String()); found {
		return version.Major, version.Minor, nil
	}

	response, err := request.Do(ctx)
	switch {
	case err != nil:
		return "", "", err
	case response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusOK:
		return "", "", response.Error()
	case response.StatusCode() == http.StatusNotFound:
		return defaultMajor, defaultMinor, nil
	}

	var version = new(resources.Version)
	err = response.ParseResponse(version)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse version response: %w", err)
	}

	saveInCache(cacheKey, version)
	return version.Major, version.Minor, nil
}

func getFromCache(key string) (*resources.Version, bool) {
	version, found := cachedVersions[key]
	return version, found
}

func saveInCache(key string, version *resources.Version) {
	cachedVersions[key] = version
}

func versionCompare(major, minor string, compareMajor, compareMinor int) bool {
	majorInt, _ := strconv.Atoi(major)

	switch {
	case majorInt < compareMajor:
		return false
	case majorInt > compareMajor:
		return true
	}

	minorInt, _ := strconv.Atoi(minor)
	return minorInt >= compareMinor
}
