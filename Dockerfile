# Copyright 2020 Coinbase, Inc.
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
FROM ubuntu:18.04 as golang-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

RUN apt-get update && apt-get install -y curl make gcc g++ git
ENV GOLANG_VERSION 1.15.5
ENV GOLANG_DOWNLOAD_SHA256 9a58494e8da722c3aef248c9227b0e9c528c7318309827780f16220998180a0d
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

# VERSION: go-ethereum v.1.9.24
RUN git clone https://github.com/ethereum/go-ethereum \
  && cd go-ethereum \
  && git checkout cc05b050df5f88e80bb26aaf6d2f339c49c2d702

RUN cd go-ethereum \
  && make geth

RUN mv go-ethereum/build/bin/geth /app/geth \
  && rm -rf go-ethereum

# Compile rosetta-ethereum
FROM golang-builder as rosetta-builder

# Use native remote build context to build in any directory
COPY . src 
RUN cd src \
  && go build

RUN mv src/rosetta-ethereum /app/rosetta-ethereum \
  && mkdir /app/ethereum \
  && mv src/ethereum/call_tracer.js /app/ethereum/call_tracer.js \
  && mv src/ethereum/geth.toml /app/ethereum/geth.toml \
  && rm -rf src 

## Build Final Image
FROM ubuntu:18.04

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from geth-builder
COPY --from=geth-builder /app/geth /app/geth

# Copy binary from rosetta-builder
COPY --from=rosetta-builder /app/ethereum /app/ethereum
COPY --from=rosetta-builder /app/rosetta-ethereum /app/rosetta-ethereum

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

CMD ["/app/rosetta-ethereum", "run"]
