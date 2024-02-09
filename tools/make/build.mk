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

##@ Go Builds Goals

.PHONY: build
build:

# if not already installed in the system install a pinned version in tools folder
GORELEASER_PATH:= $(shell command -v goreleaser 2> /dev/null)
ifndef GORELEASER_PATH
GORELEASER_PATH:= $(TOOLS_BIN)/goreleaser
endif

ifeq ($(IS_LIBRARY), 1)

BUILD_DATE:= $(shell date -u "+%Y-%m-%d")
GO_LDFLAGS+= -s -w

ifdef VERSION_MODULE_NAME
GO_LDFLAGS+= -X $(VERSION_MODULE_NAME).Version=$(VERSION)
GO_LDFLAGS+= -X $(VERSION_MODULE_NAME).BuildDate=$(BUILD_DATE)
endif

.PHONY: go/build/%
go/build/%:
	$(eval OS:= $(word 1,$(subst /, ,$*)))
	$(eval ARCH:= $(word 2,$(subst /, ,$*)))
	$(eval ARM:= $(word 3,$(subst /, ,$*)))
	$(info Building image for $(OS) $(ARCH) $(ARM))

	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) CGO_ENABLED=0 go build -trimpath \
		-ldflags "$(GO_LDFLAGS)" $(BUILD_PATH)

else

.PHONY: go/build/%
go/build/%:
	$(eval OS:= $(word 1,$(subst /, ,$*)))
	$(eval ARCH:= $(word 2,$(subst /, ,$*)))
	$(eval ARM:= $(word 3,$(subst /, ,$*)))
	$(info Building image for $(OS) $(ARCH) $(ARM))

	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) $(GORELEASER_PATH) build \
		--single-target --snapshot --clean --config=.goreleaser.yaml

.PHONY: go/build/multiarch
go/build/multiarch:
	$(GORELEASER_PATH) build --snapshot --clean --config=.goreleaser.yaml

.PHONY: build-deps
build-deps:

build-deps: $(GORELEASER_PATH)

build: build-deps

.PHONY: build-multiarch
build-multiarch: $(GORELEASER_PATH) go/build/multiarch

endif

.PHONY: build
build: go/build/$(GOOS)/$(GOARCH)/$(GOARM)

$(TOOLS_BIN)/goreleaser: $(TOOLS_DIR)/GORELEASER_VERSION
	$(eval GORELEASER_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing goreleaser $(GORELEASER_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)
