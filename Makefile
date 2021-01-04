GOOS		:= $(shell go env GOOS)
GOARCH		:= $(shell go env GOARCH)
PACKAGES	:= $(shell go list ./... | grep -vE '^keyman/cmd/')

TARGET		:= $(foreach n,$(wildcard cmd/*),$(addprefix bin/,$(notdir $n)))

# https://github.com/golang/go/issues/26492#issuecomment-435462350
ifeq ($(GOOS),windows)
TARGET		:= $(foreach t,$(TARGET),$(addsuffix .exe,$t))
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
GO.test		:= go test -race -covermode=atomic

.PHONY: all clean test

all: $(TARGET)

clean:
	$(RM) $(TARGET)

test: FORCE
	$(GO.test) -coverprofile=coverage.txt $(PACKAGES)

bin/%: FORCE
	$(GO.build) -o $@ ./cmd/$(notdir $(basename $@))

FORCE:
