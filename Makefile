NAME		:= keyman
VERSION		:= 0.0.0

GOOS		:= $(shell go env GOOS)
GOARCH		:= $(shell go env GOARCH)

TARGET		:= build/$(NAME)
LDFLAGS		:= -w
EXTLDFLAGS	:= -X main.version=$(VERSION)
TAGS		:=

ifeq ($(GOOS),windows)
TARGET		:= $(TARGET).exe
LDFLAGS		:= $(LDFLAGS) -H=windowsgui
EXTLDFLAGS	:= $(EXTLDFLAGS) -static
endif

ifneq (,$(filter $(GOOS),linux freebsd netbsd openbsd dragonfly))
TAGS		:= $(TAGS) netgo
EXTLDFLAGS	:= $(EXTLDFLAGS) -static
endif

ifeq ($(GOOS),darwin)
TAGS		:= $(TAGS) netgo
LDFLAGS		:= $(LDFLAGS) -s
EXTLDFLAGS	:= $(EXTLDFLAGS) -sectcreate __TEXT __info_plist Info.plist
endif

ifeq ($(GOOS),android)
LDFLAGS		:= $(LDFLAGS) -s
endif

GO.build	:= GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -tags '$(TAGS)' -ldflags '$(LDFLAGS) -extldflags "$(EXTLDFLAGS)"'
GO.test		:= go test -v --cover

.PHONY: default
default: build

.PHONY: build
build:
	$(GO.build) -o $(TARGET) ./cmd/$(NAME)

.PHONY: test
test:
	$(GO.test) ./...

.PHONY: clean
clean:
	$(RM) $(TARGET)
