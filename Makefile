GOOS ?= linux
GOARCH ?= amd64

PROJECT_DIR ?= $(shell pwd)

modC:
	@docker volume create golang-mod-cache
.PHONY: modC

install:
	@echo "+$@"
	@go install ./cmd/jwac/
.PHONY: install

build-jwac:
	@echo "+ $@ ${GOOS}"
	@CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/jwac/${GOOS}-${GOARCH}/jwac ./cmd/jwac/
.PHONY: build-jwac

build-jwac-all-platform:
	@GOOS=darwin make build-jwac
	@GOOS=linux make build-jwac
.PHONY: build-jwac-all-platform

arch-all:
	@zip -r arch/bin.zip bin/
.PHONY: arch-jwac-all-platform

build-jwac-tray:
	@echo "+ $@ ${GOOS}"
	@go build -o bin/jwac-tray/${GOOS}-${GOARCH}/jwac-tray ./cmd/jwac-tray/
.PHONY: build-jwac-tray

build-all: build-jwac-all-platform
	@GOOS=darwin make build-jwac-tray
.PHONY: build-jwac-tray

lint: lint-check
	@echo "+ $@"
	golangci-lint run --enable-all --skip-dirs vendor,version,gen ./pkg/filedb
.PHONY: lint

lint-check:
	@echo "+ $@"
	@if [ "`golangci-lint --version`" != "golangci-lint has version 1.33.0 built from b90551c on 2020-11-23T06:55:18Z" ]; then \
            echo "Not expected version of golangci-lint, expected 1.33.0"; \
            exit 1; \
        fi
.PHONY: lint-check

mockgen: mockgen-check
	@echo "+ $@"
	mockgen -source=pkg/filedb/json.go -destination=pkg/filedb/json_mock_test.go -package=filedb
.PHONY: lint

mockgen-check:
	@echo "+ $@"
	@if [ "`mockgen -version`" != "v1.4.4" ]; then \
            echo "Not expected version of mokcgen, expected v1.4.4"; \
            exit 1; \
        fi
.PHONY: mockgen-check
