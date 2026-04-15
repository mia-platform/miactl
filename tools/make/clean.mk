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

.PHONY: clean
clean:

.PHONY: clean/coverage
clean: clean/coverage
clean/coverage:
	$(info Clean coverage file...)
ifeq ($(OS),Windows_NT)
	if exist coverage.txt del /f /q coverage.txt
else
	rm -f coverage.txt
endif

.PHONY: clean/bin
clean: clean/bin
clean/bin:
	$(info Clean artifacts files...)
	$(call RM,$(OUTPUT_DIR))

.PHONY: clean/tools
clean/tools:
	$(info Clean tools folder...)
ifeq ($(OS),Windows_NT)
	if exist "$(subst /,\,$(TOOLS_BIN))\k8s" icacls "$(subst /,\,$(TOOLS_BIN))\k8s\*" /grant Everyone:F 2>NUL
else
	[ -d $(TOOLS_BIN)/k8s ] && chmod +w $(TOOLS_BIN)/k8s/* || true
endif
	$(call RM,$(TOOLS_BIN))

.PHONY: clean/go
clean/go:
	$(info Clean golang cache...)
	go clean -cache

.PHONY: clean-all
clean-all: clean clean/tools clean/go
