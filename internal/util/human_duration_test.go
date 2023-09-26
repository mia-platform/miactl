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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHumanDuration(t *testing.T) {
	testCases := []struct {
		duration       time.Duration
		expectedString string
	}{
		{duration: time.Second, expectedString: "1s"},
		{duration: 70 * time.Second, expectedString: "70s"},
		{duration: 190 * time.Second, expectedString: "3m10s"},
		{duration: 70 * time.Minute, expectedString: "70m"},
		{duration: 47 * time.Hour, expectedString: "47h"},
		{duration: 49 * time.Hour, expectedString: "2d1h"},
		{duration: (8*24 + 2) * time.Hour, expectedString: "8d"},
		{duration: (365*2*24 + 25) * time.Hour, expectedString: "2y1d"},
		{duration: (365*8*24 + 2) * time.Hour, expectedString: "8y"},
	}

	for _, test := range testCases {
		assert.Equal(t, test.expectedString, HumanDuration(test.duration))
	}
}

func TestHumanDurationBoundaries(t *testing.T) {
	testCases := []struct {
		duration       time.Duration
		expectedString string
	}{
		{duration: 0, expectedString: "0s"},
		{duration: time.Second - time.Millisecond, expectedString: "1s"},
		{duration: 2*time.Minute - time.Millisecond, expectedString: "120s"},
		{duration: 2 * time.Minute, expectedString: "2m0s"},
		{duration: 2*time.Minute + time.Second, expectedString: "2m1s"},
		{duration: 10*time.Minute - time.Millisecond, expectedString: "10m0s"},
		{duration: 10 * time.Minute, expectedString: "10m"},
		{duration: 10*time.Minute + time.Second, expectedString: "10m"},
		{duration: 3*time.Hour - time.Millisecond, expectedString: "180m"},
		{duration: 3 * time.Hour, expectedString: "3h0m"},
		{duration: 3*time.Hour + time.Minute, expectedString: "3h1m"},
		{duration: 8*time.Hour - time.Millisecond, expectedString: "8h0m"},
		{duration: 8 * time.Hour, expectedString: "8h"},
		{duration: 8*time.Hour + 59*time.Minute, expectedString: "8h"},
		{duration: 2*24*time.Hour - time.Millisecond, expectedString: "48h"},
		{duration: 2 * 24 * time.Hour, expectedString: "2d"},
		{duration: 2*24*time.Hour + time.Hour, expectedString: "2d1h"},
		{duration: 8*24*time.Hour - time.Millisecond, expectedString: "8d"},
		{duration: 8 * 24 * time.Hour, expectedString: "8d"},
		{duration: 8*24*time.Hour + 23*time.Hour, expectedString: "8d"},
		{duration: 2*365*24*time.Hour - time.Millisecond, expectedString: "2y"},
		{duration: 2 * 365 * 24 * time.Hour, expectedString: "2y"},
		{duration: 2*365*24*time.Hour + 23*time.Hour, expectedString: "2y"},
		{duration: 2*365*24*time.Hour + 23*time.Hour + 59*time.Minute, expectedString: "2y1d"},
		{duration: 2*365*24*time.Hour + 24*time.Hour - time.Millisecond, expectedString: "2y1d"},
		{duration: 2*365*24*time.Hour + 24*time.Hour, expectedString: "2y1d"},
		{duration: 3 * 365 * 24 * time.Hour, expectedString: "3y"},
		{duration: 4 * 365 * 24 * time.Hour, expectedString: "4y"},
		{duration: 5 * 365 * 24 * time.Hour, expectedString: "5y"},
		{duration: 6 * 365 * 24 * time.Hour, expectedString: "6y"},
		{duration: 7 * 365 * 24 * time.Hour, expectedString: "7y"},
		{duration: 8*365*24*time.Hour - time.Millisecond, expectedString: "8y"},
		{duration: 8 * 365 * 24 * time.Hour, expectedString: "8y"},
		{duration: 8*365*24*time.Hour + 364*24*time.Hour, expectedString: "8y364d"},
		{duration: 9 * 365 * 24 * time.Hour, expectedString: "9y"},
	}

	for _, test := range testCases {
		assert.Equal(t, test.expectedString, HumanDuration(test.duration))
	}
}
