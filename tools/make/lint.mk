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

##@ Lint Goals

GOLANGCI_LINT_MODE?= colored-line-number

# if not already installed in the system install a pinned version in tools folder
GOLANGCI_PATH:= $(shell command -v golangci-lint 2> /dev/null)
ifndef GOLANGCI_PATH
	GOLANGCI_PATH:=$(TOOLS_BIN)/golangci-lint
endif

.PHONY: lint
lint:

.PHONY: lint-deps
lint-deps:

.PHONY: golangci-lint
lint: golangci-lint
golangci-lint: $(GOLANGCI_PATH)
	$(info Running golangci-lint with .golangci.yaml config file...)
	$(GOLANGCI_PATH) run --out-format=$(GOLANGCI_LINT_MODE) --config=.golangci.yaml

lint-deps: $(GOLANGCI_PATH)
$(GOLANGCI_PATH): $(TOOLS_DIR)/GOLANGCI_LINT_VERSION
	$(eval GOLANGCI_LINT_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing golangci-lint $(GOLANGCI_LINT_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: gomod-lint
lint: gomod-lint
gomod-lint:
	$(info Running go mod tidy)
# Always keep this version to latest -1 version of Go
	go mod tidy -compat=1.18

.PHONY: ci-lint
ci-lint: lint
# Block the lint during ci if the go.mod and go.sum will be changed by go mod tidy
	git diff --exit-code go.mod;
	git diff --exit-code go.sum;
