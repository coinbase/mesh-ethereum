.PHONY: deps build run lint run-mainnet-online run-mainnet-offline run-testnet-online \
	run-testnet-offline check-comments add-license check-license shorten-lines \
	spellcheck salus build-local format check-format update-tracer test coverage coverage-local \
	update-bootstrap-balances mocks

ADDLICENSE_INSTALL=go install github.com/google/addlicense@latest
ADDLICENSE_CMD=addlicense
ADDLICENCE_SCRIPT=${ADDLICENSE_CMD} -c "Coinbase, Inc." -l "apache" -v
SPELLCHECK_CMD=go run github.com/client9/misspell/cmd/misspell
GOLINES_INSTALL=go install github.com/segmentio/golines@latest
GOLINES_CMD=golines
GOLINT_INSTALL=go get golang.org/x/lint/golint
GOLINT_CMD=golint
GOVERALLS_INSTALL=go install github.com/mattn/goveralls@latest
GOVERALLS_CMD=goveralls
GOIMPORTS_CMD=go run golang.org/x/tools/cmd/goimports
GO_PACKAGES=./services/... ./cmd/... ./configuration/... ./ethereum/... 
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")
TEST_SCRIPT=go test ${GO_PACKAGES}
LINT_SETTINGS=golint,misspell,gocyclo,gocritic,whitespace,goconst,gocognit,bodyclose,unconvert,lll,unparam
PWD=$(shell pwd)
NOFILE=100000

deps:
	go get ./...

test:
	${TEST_SCRIPT}

build-geth:
	docker build -t rosetta-ethereum:latest .

run-geth-mainnet:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "NETWORK=MAINNET" -p 8545:8545 -p 30303:30303 rosetta-ethereum:latest

run-geth-testnet:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "NETWORK=TESTNET" -p 8545:8545 -p 30303:30303 rosetta-ethereum:latest

run-rosetta:
	MODE=ONLINE NETWORK=MAINNET PORT=8080 FILTER=false go run *.go

run-rosetta-offline:
	MODE=OFFLINE NETWORK=MAINNET PORT=8081 FILTER=false go run *.go

local-geth:
	cd ../../.github/actions/geth; docker-compose run --rm geth ./scripts/init.sh;
	cd ../../.github/actions/geth; docker-compose up geth
