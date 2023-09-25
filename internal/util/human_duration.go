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
	"fmt"
	"time"
)

// adapted from https://github.com/kubernetes/apimachinery/blob/master/pkg/util/duration/duration.go#L48
// HumanDuration returns a succinct representation of the provided duration with limited precision for
// consumption by humans. It provides ~2-3 significant figures of duration.
func HumanDuration(d time.Duration) string {
	var durationString string

	switch {
	case d < time.Minute*2: // show seconds duration until 2 minutes
		convertedDuration := d.Round(time.Second) / time.Second
		durationString = fmt.Sprintf("%ds", convertedDuration)
	case d < time.Minute*10: // show minutes and seconds duration until 10 minutes
		convertedDuration := d.Round(time.Second)
		durationString = fmt.Sprint(convertedDuration)
	case d < time.Hour*3: // show minutes duration until 3 hours
		convertedDuration := d.Round(time.Second) / time.Minute
		durationString = fmt.Sprintf("%dm", convertedDuration)
	case d < time.Hour*8: // show hours and minutes duration until 8 hours
		convertedDuration := d.Round(time.Minute) / time.Minute
		durationString = fmt.Sprintf("%dh%dm", convertedDuration/60, convertedDuration%60)
	case d < time.Hour*48: // show hours duration until 2 days
		convertedDuration := d.Round(time.Minute) / time.Hour
		durationString = fmt.Sprintf("%dh", convertedDuration)
	case d < time.Hour*192: // show days and hours duration until ~ 8 days (24 h * 8 days = 192 hours)
		convertedDuration := d.Round(time.Minute) / time.Hour
		residualHours := convertedDuration % 24
		if residualHours == 0 {
			durationString = fmt.Sprintf("%dd", convertedDuration/24)
		} else {
			durationString = fmt.Sprintf("%dd%dh", convertedDuration/24, residualHours)
		}
	case d < time.Hour*8760: // show days duration until ~ 1 year (24 h * 365 days = 8760 hours)
		convertedDuration := d.Round(time.Hour) / time.Hour
		durationString = fmt.Sprintf("%dd", convertedDuration/24)
	default: // show days and years duration after the first year
		convertedDuration := (d.Round(time.Hour) / time.Hour) / 24
		residualDays := convertedDuration % 365
		if residualDays == 0 {
			durationString = fmt.Sprintf("%dy", convertedDuration/365)
		} else {
			durationString = fmt.Sprintf("%dy%dd", convertedDuration/365, residualDays)
		}
	}

	return durationString
}
