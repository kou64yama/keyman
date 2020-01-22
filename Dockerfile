FROM golang:1.13 AS builder

COPY . /workspace
WORKDIR /workspace

RUN make

FROM scratch

COPY --from=builder /workspace/bin/* /usr/local/bin/
ENTRYPOINT [ "keyman" ]
