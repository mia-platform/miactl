# Copyright Mia srl
# SPDX-License-Identifier: Apache-2.0

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#    http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

##@ Release Goals

SNAPSHOT_RELEASE?= 1
GORELEASER_SNAPSHOT:=

ifeq ($(SNAPSHOT_RELEASE), 1)
GORELEASER_SNAPSHOT=--snapshot
endif

.PHONY: goreleaser/release
goreleaser/release:
	$(GORELEASER_PATH) release $(GORELEASER_SNAPSHOT) --clean --config=.goreleaser.yaml

goreleaser/check:
	$(GORELEASER_PATH) check --config=.goreleaser.yaml

.PHONY: release-deps
release-deps: $(GORELEASER_PATH)

.PHONY: ci-release
ci-release: release-deps goreleaser/release

.PHONY: goreleaser-check
goreleaser-check: release-deps goreleaser/check
