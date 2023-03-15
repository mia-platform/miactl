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

package httphandler

import "fmt"

type mockValidToken struct{}
type mockExpiredToken struct{}
type mockFailAuth struct{}
type mockFailRefresh struct{}

func (a *mockValidToken) authenticate() (string, error) {
	return "valid_token", nil
}

func (a *mockExpiredToken) authenticate() (string, error) {
	if testToken == "" {
		testToken = "expired_token"
	} else {
		testToken = "valid_token"
	}
	return testToken, nil
}

func (a *mockFailAuth) authenticate() (string, error) {
	return "", fmt.Errorf("authentication failed")
}

func (a *mockFailRefresh) authenticate() (string, error) {
	if testToken == "" {
		testToken = "expired_token"
		return testToken, nil
	}
	return "", fmt.Errorf("authentication failed")
}
