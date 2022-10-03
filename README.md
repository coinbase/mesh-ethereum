<!-- HTML Graphic image –>
<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Rosetta" src="https://www.rosetta-api.org/img/rosetta_header.png">
  </a>
</p>
<!-- HTML h3 Title and Description -->
<h3 align="center">
   Rosetta Ethereum
</h3>
<p align="center">
This repository contains a sample implementation of Rosetta API for the Ethereum blockchain.
</p>
<!-- Badges -->
<p align="center">
  <a href="https://actions-badge.atrox.dev/coinbase/rosetta-ethereum/goto?ref=master"><img alt="Build Status" src="https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fcoinbase%2Frosetta-ethereum%2Fbadge%3Fref%3Dmaster&style=popout" /></a>
  <!-- <a href="https://circleci.com/gh/coinbase/rosetta-ethereum/tree/master"><img src="https://circleci.com/gh/coinbase/rosetta-ethereum/tree/master.svg?style=shield" /></a> -->
  <a href="https://coveralls.io/github/coinbase/rosetta-ethereum"><img src="https://coveralls.io/repos/github/coinbase/rosetta-ethereum/badge.svg" /></a>
  <a href="https://goreportcard.com/report/github.com/coinbase/rosetta-ethereum"><img src="https://goreportcard.com/badge/github.com/coinbase/rosetta-ethereum" /></a>
  <a href="https://github.com/coinbase/rosetta-ethereum/blob/master/LICENSE.txt"><img src="https://img.shields.io/github/license/coinbase/rosetta-ethereum.svg" /></a>
  <a href="https://pkg.go.dev/github.com/coinbase/rosetta-ethereum?tab=overview"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=shield" /></a>
</p>
<!-- Rosetta Tagline -->
<p align="center">
Build once.
Integrate your blockchain everywhere.
</p>
<!-- Release cycle information -->

<p align="center"><b>
ROSETTA-ETHEREUM IS CONSIDERED <a href="https://en.wikipedia.org/wiki/Software_release_life_cycle#Alpha">ALPHA SOFTWARE</a>.
USE AT YOUR OWN RISK.
</b></p>

<!-- Short License Info -->
<p align="center">This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).</p>

<!-- Overview -->
## Overview

The `rosetta-ethereum` repository provides an implementation sample of the Rosetta API for Ethereum in Golang. We created this repository for developers of Ethereum-like (a.k.a., account-based) blockchains, who may find it easier to fork this implementation sample than write one from scratch.

[Rosetta](https://www.rosetta-api.org/docs/welcome.html) is an open-source specification and set of tools that makes integrating with blockchains simpler, faster, and more reliable. The Rosetta API is specified in the [OpenAPI 3.0 format](https://www.openapis.org).

You can craft requests and responses with auto-generated code using [Swagger Codegen](https://swagger.io/tools/swagger-codegen) or [OpenAPI Generator](https://openapi-generator.tech). These requests and responses must be human-readable (easy to debug and understand), and able to be used in servers and browsers.

Jump to:

* [How to Use This Repo](#How-to-Use-This-Repo)
* [Docker Deployment](#Docker-Deployment)
* [Testing](#Testing)
* [Documentation](#Documentation)
* [Related Projects](#Related-Projects)
<!-- h2 How to Use This Repo -->
## How to Use This Repo

1. Fork the repo.
2. Start playing with the code.
3. [Deploy a node in Docker](#docker-deployment) to begin testing.
<!-- h3 System Requirements -->
### System Requirements

RAM: 16 MB minimum

We tested `rosetta-ethereum` on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5).
This instance type has 8 vCPU and 16 GB of RAM. If you use a computer with less than 16 GB of RAM, it is possible that `rosetta-ethereum` will exit with an OOM error.
<!-- h3 Network Settings -->
### Network Settings

To increase the load `rosetta-ethereum` can handle, we recommend these actions:

1. Tune your OS settings to allow for more connections. On a linux-based OS, run the following commands to do so ([source](http://www.tweaked.io/guide/kernel)):

```text
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_max_syn_backlog=10000
sysctl -w net.core.somaxconn=10000
sysctl -p (when done)
```
_We have not tested `rosetta-ethereum` with `net.ipv4.tcp_tw_recycle` and do not recommend enabling it._

2. Modify your open file settings to `100000`. You can do this on a linux-based OS with the command: `ulimit -n 100000`.

### Memory-Mapped Files
`rosetta-ethereum` uses [memory-mapped files](https://en.wikipedia.org/wiki/Memory-mapped_file) to persist data in the `indexer`. As a result, you **must** run `rosetta-ethereum` on a 64-bit architecture (the virtual address space easily exceeds 100s of GBs).

If you receive a kernel OOM, you may need to increase the allocated size of swap space on your OS. There is a great tutorial for how to do this on Linux [here](https://linuxize.com/post/create-a-linux-swap-file/).
<!-- h2 Configuration -->
<!-- ## Configuration -->
<!-- h2 Features -->
## Features

* Comprehensive tracking of all ETH balance changes
* Stateless, offline, curve-based transaction construction (with address checksum validation)
* Atomic balance lookups using go-ethereum's GraphQL Endpoint
* Idempotent access to all transaction traces and receipts
<!-- h2 Development -->
## Development

Helpful commands for development:

### Install dependencies
```
make deps
```

### Run Tests
```
make test
```

### Lint the Source Code
```
make lint
```

### Security Check
```
make salus
```

### Build a Docker Image from the Local Context
```
make build-local
```

### Generate a Coverage Report
```
make coverage-local
```
<!-- h2 Docker Deployment -->
## Docker Deployment

As specified in the [Rosetta API Principles](https://www.rosetta-api.org/docs/automated_deployment_testing.html), all Rosetta implementations must be deployable via Docker and support running via either an [`online` or `offline` mode](https://www.rosetta-api.org/docs/node_deployment.html#multiple-modes).

**YOU MUST INSTALL DOCKER FOR THE FOLLOWING INSTRUCTIONS TO WORK. YOU CAN DOWNLOAD DOCKER [HERE](https://www.docker.com/get-started).**
<!-- h2 Image Installation -->
### Image Installation

Running the following commands will create a Docker image called `rosetta-ethereum:latest`.

#### Installing from GitHub

To download the pre-built Docker image from the latest release, run:

```text
curl -sSfL https://raw.githubusercontent.com/coinbase/rosetta-ethereum/master/install.sh | sh -s
```
_Do not try to install rosetta-ethereum using GitHub Packages!_

#### Installing from Source

After cloning this repository, run:

```text
make build-local
```
<!-- h3 Configuring the Environment Variables -->
### Configuring the Environment Variables

#### Required Arguments

**`MODE`**
**Type:** `String`
**Options:** `ONLINE`, `OFFLINE`
**Default:** None

`MODE` determines if Rosetta can make outbound connections.

**`NETWORK`**
**Type:** `String`
**Options:** `MAINNET`, `ROPSTEN`, `RINKEBY`, `GOERLI` or `TESTNET`
**Default:** `ROPSTEN`, but only for backwards compatibility if you use `TESTNET`

`NETWORK` is the Ethereum network to launch or communicate with.

**`PORT`**
**Type:** `Integer`
**Options:** `8080`, any compatible port number
**Default:** None

`PORT` is the port to use for Rosetta.

#### Optional Arguments

**`GETH`**
**Type:** `String`
**Options:** A node URL
**Default:** None

`GETH` points to a remote `geth` node instead of initializing one

**`SKIP_GETH_ADMIN`**
**Type:** `Boolean`
**Options:** `TRUE`, `FALSE`
**Default:** `FALSE`

`SKIP_GETH_ADMIN` instructs Rosetta to not use the `geth` `admin` RPC calls. This is typically disabled by hosted blockchain node services.
<!-- h3 Run Docker -->
### Run Docker

Running the commands below will start a Docker container in
[detached mode](https://docs.docker.com/engine/reference/run/#detached--d), with
a data directory at `<working directory>/ethereum-data` and the Rosetta API accessible at port `8080`.

#### Example Commands

You can run these commands from the command line. If you cloned the repository, you can use the `make` commands shown after the examples.

**`Mainnet:Online`**

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/ethereum-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-mainnet-online
```

**`Mainnet:Online`** (Remote)

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "GETH=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-mainnet-remote geth=<NODE URL>
```

**`Mainnet:Offline`**

Uncloned repo:
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-mainnet-offline
```

**`Testnet:Online`**

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/ethereum-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-testnet-online
// to send some funds into your testnet account you can use the following commands
geth attach http://127.0.0.1:8545
> eth.sendTransaction({from: eth.coinbase, to: "0x9C639954BC9956598Df734994378A36f73cfba0C", value: web3.toWei(50, "ether")})
```

**`Testnet:Online`** (Remote)

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "GETH=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-testnet-remote geth=<NODE URL>
```

**`Testnet:Offline`**

Uncloned repo:
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-ethereum:latest
```

Cloned repo:
```text
make run-testnet-offline
```
<!-- h2 Testing -->
## Test the Implementation with rosetta-cli

To validate `rosetta-ethereum`, [install `rosetta-cli`](https://github.com/coinbase/rosetta-cli#install)
and run one of the following commands:

* `rosetta-cli check:data --configuration-file rosetta-cli-conf/testnet/config.json` - This command validates that the Data API implementation is correct using the ethereum `testnet` node. It also ensures that the implementation does not miss any balance-changing operations.
* `rosetta-cli check:construction --configuration-file rosetta-cli-conf/testnet/config.json` - This command validates the Construction API implementation. It also verifies transaction construction, signing, and submissions to the `testnet` network.
* `rosetta-cli check:data --configuration-file rosetta-cli-conf/mainnet/config.json` - This command validates that the Data API implementation is correct using the ethereum `mainnet` node. It also ensures that the implementation does not miss any balance-changing operations.

Read the [How to Test your Rosetta Implementation](https://www.rosetta-api.org/docs/rosetta_test.html) documentation for additional details.
<!-- h2 Contributing -->
## Contributing

You may contribute to the `rosetta-ethereum` project in various ways:

* [Asking Questions](CONTRIBUTING.md/#asking-questions)
* [Providing Feedback](CONTRIBUTING.md/#providing-feedback)
* [Reporting Issues](CONTRIBUTING.md/#reporting-issues)

Read our [Contributing](CONTRIBUTING.MD) documentation for more information.

When you've finished an implementation for a blockchain, share your work in the [ecosystem category of the community site](https://community.rosetta-api.org/c/ecosystem). Platforms looking for implementations for certain blockchains will be monitoring this section of the website for high-quality implementations they can use for integration. Make sure that your implementation meets the [expectations](https://www.rosetta-api.org/docs/node_deployment.html) of any implementation.

You can also find community implementations for a variety of blockchains in the [rosetta-ecosystem](https://github.com/coinbase/rosetta-ecosystem) repository.
<!-- h2 Documentation -->
## Documentation

You can find the Rosetta API documentation at [rosetta-api.org](https://www.rosetta-api.org/docs/welcome.html).

Check out the [Getting Started](https://www.rosetta-api.org/docs/getting_started.html) section to start diving into Rosetta.

Our documentation is divided into the following sections:

* [Product Overview](https://www.rosetta-api.org/docs/welcome.html)
* [Getting Started](https://www.rosetta-api.org/docs/getting_started.html)
* [Rosetta API Spec](https://www.rosetta-api.org/docs/Reference.html)
* [Samples](https://www.rosetta-api.org/docs/reference-implementations.html)
* [Testing](https://www.rosetta-api.org/docs/rosetta_cli.html)
* [Best Practices](https://www.rosetta-api.org/docs/node_deployment.html)
* [Repositories](https://www.rosetta-api.org/docs/rosetta_specifications.html)
<!-- h2 Related Projects -->
## Related Projects

* [rosetta-geth-sdk](https://github.com/coinbase/rosetta-geth-sdk) — This SDK helps accelerate Rosetta API implementation on go-ethereum based chains.
* [rosetta-sdk-go](https://github.com/coinbase/rosetta-sdk-go) — The `rosetta-sdk-go` SDK provides a collection of packages used for interaction with the Rosetta API specification.
* [rosetta-specifications](https://github.com/coinbase/rosetta-specifications) — Much of the SDKs’ code is generated from this repository.
* [rosetta-cli](https://github.com/coinbase/rosetta-ecosystem) — Use the `rosetta-cli` tool to test your Rosetta API implementation. The tool also provides the ability to look up block contents and account balances.
<!-- h3 Other Implementation Samples -->
### Other Implementation Samples

You can find community implementations for a variety of blockchains in the [rosetta-ecosystem](https://github.com/coinbase/rosetta-ecosystem) repository, and in the [ecosystem category](https://community.rosetta-api.org/c/ecosystem) of our community site.

<!-- h2 License and Copyright -->
## License

This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).

© 2022 Coinbase
