# Copyright 2022 Coinbase, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Compile golang
FROM ubuntu:20.04 as golang-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

RUN apt-get update && apt-get install -y curl make gcc g++ git
ENV GOLANG_VERSION 1.16.8
ENV GOLANG_DOWNLOAD_SHA256 f32501aeb8b7b723bc7215f6c373abb6981bbc7e1c7b44e9f07317e1a300dce2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Compile geth
FROM golang-builder as geth-builder

# VERSION: go-ethereum v.1.10.16
RUN git clone https://github.com/ethereum/go-ethereum \
  && cd go-ethereum \
  && git checkout 20356e57b119b4e70ce47665a71964434e15200d

RUN cd go-ethereum \
  && make geth

RUN mv go-ethereum/build/bin/geth /app/geth \
  && rm -rf go-ethereum

# Compile rosetta-ethereum
FROM golang-builder as rosetta-builder

# Copy binary from geth-builder
COPY --from=geth-builder /app/geth /app/geth

# Use native remote build context to build in any directory
COPY . src

RUN mv src/geth.toml /app/geth.toml \
  && mv src/entrypoint.sh /app/entrypoint.sh \
  && rm -rf src

## Build Final Image
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from geth-builder
COPY --from=geth-builder /app/geth /app/geth

# Copy binary from rosetta-builder
COPY --from=rosetta-builder /app /app

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

EXPOSE 8545 8546 30303 30303/udp

ENTRYPOINT ["/app/entrypoint.sh"]

