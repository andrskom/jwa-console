GOOS ?= linux
GOARCH ?= amd64

install:
	@echo "+$@"
	@go install ./cmd/jwac/
.PHONY: install

build-jwac:
	@echo "+ $@ ${GOOS}"
	@CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/${GOOS}-${GOARCH}/jwac ./cmd/jwac/
.PHONY: build-jwac

build-jwac-all-platform:
	@GOOS=darwin make build-jwac
	@GOOS=linux make build-jwac
.PHONY: build-jwac-all-platform