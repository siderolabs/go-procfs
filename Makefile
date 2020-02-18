REGISTRY ?= docker.io
USERNAME ?= autonomy
SHA ?= $(shell git describe --match=none --always --abbrev=8 --dirty)
TAG ?= $(shell git describe --tag --always --dirty)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
REGISTRY_AND_USERNAME := $(REGISTRY)/$(USERNAME)
NAME := CHANGEME
IMAGE := $(REGISTRY_AND_USERNAME)/$(NAME)
MODULE := github.com/talos-systems/$(NAME)

ARTIFACTS := _out
OPERATING_SYSTEM := $(shell uname -s | tr "[:upper:]" "[:lower:]")
PKGS ?= ./...

BUILD := docker buildx build
PLATFORM ?= linux/amd64
PROGRESS ?= auto
PUSH ?= false
COMMON_ARGS := --file=Dockerfile
COMMON_ARGS += --progress=$(PROGRESS)
COMMON_ARGS += --platform=$(PLATFORM)
COMMON_ARGS += --push=$(PUSH)
COMMON_ARGS += --build-arg=SHA=$(SHA)
COMMON_ARGS += --build-arg=TAG=$(TAG)
COMMON_ARGS += --build-arg=MODULE=$(MODULE)
COMMON_ARGS += --build-arg=PKGS=$(PKGS)

all: container

# Help Menu

define HELP_MENU_HEADER
# Getting Started

To build this project, you must have the following installed:

- git
- make
- docker (19.03 or higher)
- buildx (https://github.com/docker/buildx)

## Creating a Builder Instance

The build process makes use of features not currently supported by the default
builder instance (docker driver). To create a compatible builder instance, run:

```
docker buildx create --driver docker-container --name local --buildkitd-flags --use
```

If you already have a compatible builder instance, you may use that instead.

## Artifacts

All artifacts will be output to ./$(ARTIFACTS). Images will be tagged with the
registry "$(REGISTRY)", username "$(USERNAME)", and a dynamic tag (e.g. $(IMAGE):$(TAG)).
The registry and username can be overriden by exporting REGISTRY, and USERNAME
respectively.

endef

export HELP_MENU_HEADER

help: ## This help menu.
	@echo "$$HELP_MENU_HEADER"
	@grep -E '^[a-zA-Z%_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build Abstractions

target-%: ## Builds the specified target defined in the Dockerfile. The build result will remain only in the build cache.
	@$(BUILD) --target=$* $(COMMON_ARGS) $(TARGET_ARGS) .

local-%: ## Builds the specified target defined in the Dockerfile using the local output type. The build result will be output to the specified local destination.
	@$(MAKE) target-$* TARGET_ARGS="--output=type=local,dest=$(DEST) $(TARGET_ARGS)"

docker-%: ## Builds the specified target defined in the Dockerfile using the docker output type. The build result will be loaded into docker.
	@$(MAKE) target-$* TARGET_ARGS="--tag $(IMAGE):$(TAG) $(TARGET_ARGS)"

# Artifacts

.PHONY: container
container: ## Build the container image.
	@$(MAKE) docker-$@ TARGET_ARGS="--push=$(PUSH)"

# Code Quality

.PHONY: fmt
fmt: ## Formats the source code.
	@$(MAKE) local-fmt DEST=./

lint-%: ## Runs the specified linter. Valid options are go, protobuf, and markdown (e.g. lint-go).
	@$(MAKE) target-lint-$*

lint: ## Runs linters on go, protobuf, and markdown file types.
	@$(MAKE) lint-go lint-markdown

# Tests

.PHONY: unit-tests
unit-tests: ## Performs unit tests.
	@$(MAKE) local-$@ DEST=$(ARTIFACTS)

.PHONY: unit-tests-race
unit-tests-race: ## Performs unit tests with race detection enabled.
	@$(MAKE) target-$@

# Utilities

.PHONY: login
login: ## Logs in to the configured container registry.
	@docker login --username "$(DOCKER_USERNAME)" --password "$(DOCKER_PASSWORD)" $(REGISTRY)

.PHONY: clean
clean: ## Cleans up all artifacts.
	@-rm -rf $(ARTIFACTS)
