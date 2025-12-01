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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/mia-platform/miactl/internal/clioptions"
)

// Version is dynamically set by the ci or overridden by the Makefile.
var Version = ""

// BuildDate is dynamically set at build time by the cli or overridden in the Makefile.
var BuildDate = "" // YYYY-MM-DD

func VersionCmd(_ *clioptions.CLIOptions) *cobra.Command {
	versionOutput := versionFormat(Version, BuildDate)
	// Version subcommand
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show miactl version",
		Long:  "Show miactl version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(versionOutput)
		},
	}

	return cmd
}

// versionFormat return the version string nicely formatted
func versionFormat(version, buildDate string) string {
	if buildDate != "" {
		version = fmt.Sprintf("%s (%s)", version, buildDate)
	}

	osCommand := os.Args[0]
	version = fmt.Sprintf("%s version: %s", filepath.Base(osCommand), version)
	// don't return GoVersion during a test run for consistent test output
	if flag.Lookup("test.v") != nil {
		return version
	}
	return fmt.Sprintf("%s, Go Version: %s (%s/%s)", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
