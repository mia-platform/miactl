# Copyright Mia srl
# SPDX-License-Identifier: Apache-2.0

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DEBUG_MAKEFILE?=
ifeq ($(DEBUG_MAKEFILE),1)
$(warning ***** executing goal(s) "$(MAKECMDGOALS)")
$(warning ***** $(shell date))
else
# If we're not debugging the Makefile, always hide the commands inside the goals
MAKEFLAGS+= -s
endif

# It's necessary to set this because some environments don't link sh -> bash.
# Using env is more portable than setting the path directly
SHELL:= /usr/bin/env bash

.EXPORT_ALL_VARIABLES:

.SUFFIXES:

## Set all variables
ifeq ($(origin PROJECT_DIR),undefined)
PROJECT_DIR:= $(abspath $(shell pwd -P))
endif

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR:= $(PROJECT_DIR)/bin
endif

ifeq ($(origin TOOLS_DIR),undefined)
TOOLS_DIR:= $(PROJECT_DIR)/tools
endif

ifeq ($(origin TOOLS_BIN),undefined)
TOOLS_BIN:= $(TOOLS_DIR)/bin
endif

# Set here the name of the package you want to build
CMDNAME:= miactl
BUILD_PATH:= .
CONFORMANCE_TEST_PATH:= $(PROJECT_DIR)/tests/e2e
IS_LIBRARY:=

# enable modules
GO111MODULE:= on
GOOS:= $(shell go env GOOS)
GOARCH:= $(shell go env GOARCH)
GOARM:= $(shell go env GOARM)

## Build Variables
GIT_REV:= $(shell git rev-parse --short HEAD 2>/dev/null)
VERSION:= $(shell git describe --tags --exact-match 2>/dev/null || (echo $(GIT_REV) | cut -c1-12))
# insert here the go module where to add the version metadata
VERSION_MODULE_NAME:= github.com/mia-platform/miactl/internal/cmd

# supported platforms for container creation, these are a subset of the supported
# platforms of the base image.
# Or if you start from scratch the platforms you want to support in your image
# This link contains the rules on how the strings must be formed https://github.com/containerd/containerd/blob/v1.4.3/platforms/platforms.go#L63
SUPPORTED_PLATFORMS:= linux/386 linux/amd64 linux/arm64 linux/arm/v6 linux/arm/v7
# Default platform for which building the docker image (darwin can run linux images for the same arch)
# as SUPPORTED_PLATFORMS it highly depends on which platform are supported by the base image
DEFAULT_DOCKER_PLATFORM:= linux/$(GOARCH)/$(GOARM)
# List of one or more container registries for tagging the resulting docker images
CONTAINER_REGISTRIES:= docker.io/miaplatform ghcr.io/mia-platform
# The description used on the org.opencontainers.description label of the container
DESCRIPTION:= Mia Platform Cli for Console
# The vendor name used on the org.opencontainers.image.vendor label of the container
VENDOR_NAME:= Mia s.r.l.
# The license used on the org.opencontainers.image.license label of the container
LICENSE:= Apache-2.0
# The documentation url used on the org.opencontainers.image.documentation label of the container
DOCUMENTATION_URL:= https://docs.mia-platform.eu
# The source url used on the org.opencontainers.image.source label of the container
SOURCE_URL:= https://github.com/mia-platform/miactl
BUILDX_CONTEXT?= miactl-build-context

# Add additional targets that you want to run when calling make without arguments
.PHONY: all
all: lint test

## Includes
include tools/make/clean.mk
include tools/make/lint.mk
include tools/make/test.mk
include tools/make/generate.mk
include tools/make/build.mk
include tools/make/container.mk
include tools/make/release.mk

# Uncomment the correct test suite to run during CI
.PHONY: ci
ci: test-coverage
# ci: test-integration-coverage

### Put your custom import, define or goals under here ###

generate-deps: $(TOOLS_BIN)/deepcopy-gen
$(TOOLS_BIN)/deepcopy-gen: $(TOOLS_DIR)/DEEPCOPY_GEN_VERSION
	$(eval DEEPCOPY_GEN_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing deepcopy-gen $(DEEPCOPY_GEN_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(DEEPCOPY_GEN_VERSION)

BUILD_ALPHA?=false

.PHONY: build-alpha
build-alpha:
	BUILD_ALPHA=true $(MAKE) build
