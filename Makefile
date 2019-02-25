SOURCE	= $(shell find . -type d -name vendor -prune -o -type f -name '*.go' -not -name '*_test.go' -print)
PACKAGE	= ./cmd/keyman
OUTPUT	= dist

GO	= go
GOX	= gox
GOOS	= $(shell go env GOOS)
GOARCH	= $(shell go env GOARCH)
GO_LDFLAGS	= -ldflags '-extldflags "-static"'
GOX_OS_ARCH	= \
	darwin/386 \
	darwin/amd64 \
	linux/386 \
	linux/amd64 \
	linux/arm \
	freebsd/386 \
	freebsd/amd64 \
	openbsd/386 \
	openbsd/amd64 \
	windows/386 \
	windows/amd64 \
	freebsd/arm \
	netbsd/386 \
	netbsd/amd64 \
	netbsd/arm

all: build
	file $(wildcard $(OUTPUT)/*)

clean:
	$(RM) $(wildcard $(OUTPUT)/*)

lint:
	$(GO) vet ./...

test: lint
	$(GO) test -v -cover ./...

build: test
	CGO_ENABLED=0 $(GOX) $(GO_LDFLAGS) -osarch="$(GOX_OS_ARCH)" -output $(OUTPUT)/{{.Dir}}_{{.OS}}_{{.Arch}} $(PACKAGE)

.PHONY: all clean lint test build
