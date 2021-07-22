// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/coinbase/rosetta-ethereum/ethereum"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/params"
)

// Mode is the setting that determines if
// the implementation is "online" or "offline".
type Mode string

const (
	// Online is when the implementation is permitted
	// to make outbound connections.
	Online Mode = "ONLINE"

	// Offline is when the implementation is not permitted
	// to make outbound connections.
	Offline Mode = "OFFLINE"

	// Mainnet is the Ethereum Mainnet.
	Mainnet string = "MAINNET"

	// Ropsten is the Ethereum Ropsten testnet.
	Ropsten string = "ROPSTEN"

	// Rinkeby is the Ethereum Rinkeby testnet.
	Rinkeby string = "RINKEBY"

	// Goerli is the Ethereum Görli testnet.
	Goerli string = "GOERLI"

	// Testnet defaults to `Ropsten` for backwards compatibility.
	Testnet string = "TESTNET"

	// DataDirectory is the default location for all
	// persistent data.
	DataDirectory = "/data"

	// ModeEnv is the environment variable read
	// to determine mode.
	ModeEnv = "MODE"

	// NetworkEnv is the environment variable
	// read to determine network.
	NetworkEnv = "NETWORK"

	// PortEnv is the environment variable
	// read to determine the port for the Rosetta
	// implementation.
	PortEnv = "PORT"

	// GethEnv is an optional environment variable
	// used to connect rosetta-ethereum to an already
	// running geth node.
	GethEnv = "GETH"

	// DefaultGethURL is the default URL for
	// a running geth node. This is used
	// when GethEnv is not populated.
	DefaultGethURL = "http://localhost:8545"

	// SkipGethAdminEnv is an optional environment variable
	// to skip geth `admin` calls which are typically not supported
	// by hosted node services. When not set, defaults to false.
	SkipGethAdminEnv = "SKIP_GETH_ADMIN"

	// GethHeadersEnv is an optional environment variable
	// of a comma-separated list of key:value pairs to apply
	// to geth clients as headers. When not set, defaults to []
	GethHeadersEnv = "GETH_HEADERS"

	// MiddlewareVersion is the version of rosetta-ethereum.
	MiddlewareVersion = "0.0.4"
)

// Configuration determines how
type Configuration struct {
	Mode                   Mode
	Network                *types.NetworkIdentifier
	GenesisBlockIdentifier *types.BlockIdentifier
	GethURL                string
	RemoteGeth             bool
	Port                   int
	GethArguments          string
	SkipGethAdmin          bool
	GethHeaders            []*ethereum.HTTPHeader

	// Block Reward Data
	Params *params.ChainConfig
}

// LoadConfiguration attempts to create a new Configuration
// using the ENVs in the environment.
func LoadConfiguration() (*Configuration, error) {
	config := &Configuration{}

	modeValue := Mode(os.Getenv(ModeEnv))
	switch modeValue {
	case Online:
		config.Mode = Online
	case Offline:
		config.Mode = Offline
	case "":
		return nil, errors.New("MODE must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid mode", modeValue)
	}

	networkValue := os.Getenv(NetworkEnv)
	switch networkValue {
	case Mainnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: ethereum.Blockchain,
			Network:    ethereum.MainnetNetwork,
		}
		config.GenesisBlockIdentifier = ethereum.MainnetGenesisBlockIdentifier
		config.Params = params.MainnetChainConfig
		config.GethArguments = ethereum.MainnetGethArguments
	case Testnet, Ropsten:
		config.Network = &types.NetworkIdentifier{
			Blockchain: ethereum.Blockchain,
			Network:    ethereum.RopstenNetwork,
		}
		config.GenesisBlockIdentifier = ethereum.RopstenGenesisBlockIdentifier
		config.Params = params.RopstenChainConfig
		config.GethArguments = ethereum.RopstenGethArguments
	case Rinkeby:
		config.Network = &types.NetworkIdentifier{
			Blockchain: ethereum.Blockchain,
			Network:    ethereum.RinkebyNetwork,
		}
		config.GenesisBlockIdentifier = ethereum.RinkebyGenesisBlockIdentifier
		config.Params = params.RinkebyChainConfig
		config.GethArguments = ethereum.RinkebyGethArguments
	case Goerli:
		config.Network = &types.NetworkIdentifier{
			Blockchain: ethereum.Blockchain,
			Network:    ethereum.GoerliNetwork,
		}
		config.GenesisBlockIdentifier = ethereum.GoerliGenesisBlockIdentifier
		config.Params = params.GoerliChainConfig
		config.GethArguments = ethereum.GoerliGethArguments
	case "":
		return nil, errors.New("NETWORK must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid network", networkValue)
	}

	config.GethURL = DefaultGethURL
	envGethURL := os.Getenv(GethEnv)
	if len(envGethURL) > 0 {
		config.RemoteGeth = true
		config.GethURL = envGethURL
	}

	config.SkipGethAdmin = false
	envSkipGethAdmin := os.Getenv(SkipGethAdminEnv)
	if len(envSkipGethAdmin) > 0 {
		val, err := strconv.ParseBool(envSkipGethAdmin)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to parse SKIP_GETH_ADMIN %s", err, envSkipGethAdmin)
		}
		config.SkipGethAdmin = val
	}

	envGethHeaders := os.Getenv(GethHeadersEnv)
	if len(envGethHeaders) > 0 {
		headers := strings.Split(envGethHeaders, ",")
		headerKVs := make([]*ethereum.HTTPHeader, len(headers))
		for i, pair := range headers {
			kv := strings.Split(pair, ":")
			headerKVs[i] = &ethereum.HTTPHeader{
				Key:   kv[0],
				Value: kv[1],
			}
		}
		config.GethHeaders = headerKVs
	}

	portValue := os.Getenv(PortEnv)
	if len(portValue) == 0 {
		return nil, errors.New("PORT must be populated")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || len(portValue) == 0 || port <= 0 {
		return nil, fmt.Errorf("%w: unable to parse port %s", err, portValue)
	}
	config.Port = port

	return config, nil
}
