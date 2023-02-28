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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	t.Run("without correct args", func(t *testing.T) {
		_, err := executeCommand(NewRootCmd(), "completion", "not-correct-arg")
		expectedErrMessage := `invalid argument "not-correct-arg" for "miactl completion"`
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("without args", func(t *testing.T) {
		_, err := executeCommand(NewRootCmd(), "completion")
		expectedErrMessage := `accepts 1 arg(s), received 0`
		require.EqualError(t, err, expectedErrMessage)
	})

	t.Run("with fish arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "fish")
		require.Nil(t, err)
		require.Contains(t, out, "# fish completion for miactl")
	})

	t.Run("with bash arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "bash")
		require.Nil(t, err)
		require.Contains(t, out, "# bash completion for miactl")
	})

	t.Run("with zsh arg", func(t *testing.T) {
		out, err := executeCommand(NewRootCmd(), "completion", "zsh")
		require.Nil(t, err)
		require.Contains(t, out, "#compdef miactl")
	})
}
