FROM golang:1.13 AS builder

COPY . /workspace
WORKDIR /workspace

ARG VERSION=0.0.0
RUN make LDFLAGS="-X main.version=${VERSION}"

FROM scratch

COPY --from=builder /workspace/bin/* /usr/local/bin/

ENTRYPOINT [ "keyman" ]
