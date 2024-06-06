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

##@ Go Tests Goals

DEBUG_TEST?=
ifeq ($(DEBUG_TEST),1)
GO_TEST_DEBUG_FLAG:= -v
else
GO_TEST_DEBUG_FLAG:=
endif

ENVTEST_K8S_VERSION?= $(shell cat $(TOOLS_DIR)/ENVTEST_K8S_VERSION)

# needed until the linker warning on macos are fixed
ifeq (Darwin,$(shell uname -s))
MACOS_LINKER_FLAGS ?= -extldflags=-Wl,-ld_classic
endif

.PHONY: test/unit
test/unit:
	$(info Running tests...)
	go test $(GO_TEST_DEBUG_FLAG) -ldflags="$(MACOS_LINKER_FLAGS)" -race ./...

.PHONY: test/integration
test/integration:
	$(info Running integration tests...)
	go test $(GO_TEST_DEBUG_FLAG) -ldflags="$(MACOS_LINKER_FLAGS)" -tags=integration -race ./...

.PHONY: test/coverage
test/coverage:
	$(info Running tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -ldflags="$(MACOS_LINKER_FLAGS)" -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/integration/coverage
test/integration/coverage:
	$(info Running ci tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -ldflags="$(MACOS_LINKER_FLAGS)" -tags=integration -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/conformance test/conformance/setup test/conformance/teardown
test/conformance/setup:
test/conformance:
	$(info Running conformance tests...)
	go test $(GO_TEST_DEBUG_FLAG) -ldflags="$(MACOS_LINKER_FLAGS)" -tags=conformance -race -count=1 $(CONFORMANCE_TEST_PATH)
test/conformance/teardown:

test/show/coverage:
	go tool cover -func=coverage.txt

.PHONY: test
test: test/unit

.PHONY: test-coverage
test-coverage: test/coverage

.PHONY: test-integration
test-integration: $(TOOLS_BIN)/setup-envtest envtest/assets test/integration

.PHONY: test-integration-coverage
test-integration-coverage: $(TOOLS_BIN)/setup-envtest envtest/assets test/integration/coverage

.PHONY: test-conformance
test-conformance: test/conformance/setup test/conformance test/conformance/teardown

.PHONY: show-coverage
show-coverage: test-coverage test/show/coverage

.PHONY: show-integration-coverage
show-integration-coverage: test-integration-coverage test/show/coverage

$(TOOLS_BIN)/setup-envtest: $(TOOLS_DIR)/ENVTEST_VERSION
	$(eval ENVTEST_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing testenv $(ENVTEST_VERSION) bin in $(TOOLS_BIN)...)
	GOBIN=$(TOOLS_BIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)

.PHONY: envtest/assets
envtest/assets:
	$(info Setup testenv with k8s $(ENVTEST_K8S_VERSION) version...)
	$(eval KUBEBUILDER_ASSETS:= $(shell $(TOOLS_BIN)/setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(TOOLS_BIN) -p path))
