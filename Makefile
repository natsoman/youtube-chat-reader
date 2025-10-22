REGISTRY      ?= ghcr.io
VERSION       ?= local
DOCKERFILE    ?= build/Dockerfile
NAMESPACE     ?= natsoman

IMAGE := $(REGISTRY)/$(NAMESPACE)/$(IMAGE_NAME)
GOMODULES := ./apps/finder/... ./apps/reader/... ./pkg/kafka/... ./pkg/mongo/... ./pkg/otel/...

.DEFAULT_GOAL := lint
.PHONY: all test lint reader-consumer reader-worker finder build proto-gen

all: lint test reader-worker reader-consumer finder

test:
	go test -tags integration -race -count=1 $(GOMODULES)

lint:
	golangci-lint run --fix $(GOMODULES)

reader-consumer:
	make build GOTARGET=apps/reader/cmd/consumer/main.go IMAGE_NAME=reader-consumer

reader-worker:
	make build GOTARGET=apps/reader/cmd/worker/main.go IMAGE_NAME=reader-worker

finder:
	make build GOTARGET=apps/finder/cmd/job/main.go IMAGE_NAME=finder

build: # Builds local images (no push) that can be used from local k8s cluster
	@if [ -z "$(GOTARGET)" ]; then \
		echo "GOTARGET is required"; \
		exit 1; \
	fi
	@if [ -z "$(IMAGE_NAME)" ]; then \
		echo "IMAGE_NAME is required"; \
		exit 1; \
	fi
	@echo "Building $(IMAGE):$(VERSION)"
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GOTARGET=$(GOTARGET) \
		-f $(DOCKERFILE) \
		-t $(IMAGE):$(VERSION) \
		--load \
		.

proto-gen:
	protoc --go_out=. \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=. \
	  --go-grpc_opt=paths=source_relative apps/reader/internal/infra/youtube/stream_list.proto
