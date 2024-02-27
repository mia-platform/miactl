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

package authorization

import "os/exec"

func openBrowser(url string) error {
	// support different commands on linux and exec the first one found
	commands := []string{"open", "xdg-open"}
	// Look for one that exists and run it
	for _, command := range commands {
		if _, err := exec.LookPath(command); err == nil {
			cmd := exec.Command(command, url)
			return cmd.Run()
		}
	}

	return &exec.Error{Name: strings.Join(commands, ", "), Err: exec.ErrNotFound}
}
