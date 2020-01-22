GOOS		:= $(shell go env GOOS)
GOARCH		:= $(shell go env GOARCH)

PROTO		:= $(wildcard pb/*.proto)
GENERATED	:= $(PROTO:.proto=.pb.go)
OUTPUT		:= bin

ifeq ($(GOOS),windows)
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

GO.build	:= go build -tags '$(TAGS)' -ldflags '$(LDFLAGS) -extldflags "$(EXTLDFLAGS)"'

.PHONY: build proto clean
.SUFFIXES: .pb.go .proto

build: $(OUTPUT) proto
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GO.build) -o $< ./...

proto: $(GENERATED)

clean:
	$(RM) $(wildcard $(OUTPUT)/*)

$(OUTPUT):
	mkdir -p $@

pb/%.pb.go: pb/%.proto
	protoc -I pb/ $< --go_out=plugins=grpc:pb
