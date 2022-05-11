.DEFAULT_GOAL := build
TAG ?=1.0.0
ECR_REGISTRY ?=
COMMIT_SHA ?= git-$(shell git rev-parse --short HEAD)
BUILD_ID ?= "UNSET"
PKG = github.com/kaladaOpuiyo/ecr-mirror-sync 

HOST_ARCH = $(shell which go >/dev/null 2>&1 && go env GOARCH)
ARCH ?= $(HOST_ARCH)
ifeq ($(ARCH),)
    $(error mandatory variable ARCH is empty, either set it when calling the command or make sure 'go env GOARCH' works)
endif


.PHONY: build
build:  ## Build ecr-mirror-sync.
	PKG=$(PKG) \
	ARCH=$(ARCH) \
	TAG=$(TAG) \
	build/build.sh

.PHONY: image
BASE_IMAGE=golang:1.16
image: clean-image ## Build image for a particular arch.
	@echo "Building docker image ($(ARCH))..."
	@docker build \
		--no-cache \
		--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
		--build-arg VERSION="$(TAG)" \
		-t $(ECR_REGISTRY)/ecr-mirror-sync:$(TAG) .

.PHONY: clean-image
clean-image: ## Removes local image
	@echo "Removing old image $(REGISTRY)/ecr-mirror-sync:$(TAG)"
	@docker rmi -f $(ECR_REGISTRY)/ecr-mirror-sync:$(TAG) || true

.PHONY: clean-package-image
clean-package-image: ## Removes local image
	@echo "Removing old image $(ECR_REGISTRY)/ecr-mirror-sync-package:$(TAG)"
	@docker rmi -f $(ECR_REGISTRY)/ecr-mirror-sync-package:$(TAG) || true

.PHONY: test
test: 
	go test ./...
