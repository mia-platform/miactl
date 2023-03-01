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

package renderer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Ref: https://play.golang.org/p/6KBd7Dd3UJy
func TestDateFormatOutput(t *testing.T) {
	date, err := time.Parse(time.UnixDate, "Sat Mar  7 11:06:39 PST 2015")
	require.NoError(t, err)

	require.Equal(t, "07 Mar 2015 11:06 PST", FormatDate(date))
}
