GOOS ?= linux
GOARCH ?= amd64

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
