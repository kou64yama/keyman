FROM golang:1.13 AS builder

COPY . /workspace
WORKDIR /workspace

RUN make

FROM alpine:3.12

RUN addgroup -g 1000 -S keyman \
  && adduser -h /var/lib/keyman -s /sbin/nologin -G keyman -S -D -u 1000 keyman
COPY --from=builder /workspace/bin/* /usr/local/bin/

VOLUME [ "/var/lib/keyman" ]

WORKDIR /var/lib/keyman
USER keyman
ENV KEYMAN_HOME=/var/lib/keyman

ENTRYPOINT [ "keyman" ]
