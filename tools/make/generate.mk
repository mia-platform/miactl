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

##@ Deepcopy Goals

.PHONY: generate
generate:

.PHONY: generate-deps
generate-deps:

.PHONY: generate/deepcopy
generate/deepcopy:
	$(info Running deepcopy-gen...)
	$(TOOLS_BIN)/deepcopy-gen -i $(PACKAGES_TO_GENERATE) \
		-o "$(PROJECT_DIR)" -O zz_generated.deepcopy --go-header-file $(TOOLS_DIR)/boilerplate.go.txt

$(TOOLS_BIN)/deepcopy-gen: $(TOOLS_DIR)/DEEPCOPY_GEN_VERSION
	$(eval DEEPCOPY_GEN_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing deepcopy-gen $(DEEPCOPY_GEN_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(DEEPCOPY_GEN_VERSION)
generate-deps: $(TOOLS_BIN)/deepcopy-gen

.PHONY: generate-deepcopy
generate-deepcopy: $(TOOLS_BIN)/deepcopy-gen generate/deepcopy
generate: generate-deepcopy
