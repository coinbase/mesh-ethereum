// Copyright 2022 Coinbase, Inc.
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

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/coinbase/rosetta-geth-sdk/configuration"
	RosettaTypes "github.com/coinbase/rosetta-sdk-go/types"
)

// Mode is the setting that determines if
// the implementation is "online" or "offline".
type Mode string

const (
	// Blockchain is Ethereum.
	Blockchain string = "Ethereum"

	// MainnetNetwork is the value of the network
	// in MainnetNetworkIdentifier.
	MainnetNetwork string = "Mainnet"

	// RopstenNetwork is the value of the network
	// in RopstenNetworkIdentifier.
	RopstenNetwork string = "Ropsten"

	// RinkebyNetwork is the value of the network
	// in RinkebyNetworkNetworkIdentifier.
	RinkebyNetwork string = "Rinkeby"

	// GoerliNetwork is the value of the network
	// in GoerliNetworkNetworkIdentifier.
	GoerliNetwork string = "Goerli"

	// Symbol is the symbol value
	// used in Currency.
	Symbol = "ETH"

	// Decimals is the decimals value
	// used in Currency.
	Decimals = 18

	// Online is when the implementation is permitted
	// to make outbound connections.
	Online configuration.Mode = "ONLINE"

	// Offline is when the implementation is not permitted
	// to make outbound connections.
	Offline configuration.Mode = "OFFLINE"

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

	// TokenFilterEnv is the environment variable
	// read to determine if we will filter tokens
	// using our token white list
	TokenFilterEnv = "FILTER"

	// GethEnv is an optional environment variable
	// used to connect rosetta-ethereum to an already
	// running geth node.
	GethEnv = "GETH"

	// DefaultGethURL is the default URL for
	// a running geth node. This is used
	// when GethEnv is not populated.
	DefaultGethURL = "http://127.0.0.1:8546"

	// SkipGethAdminEnv is an optional environment variable
	// to skip geth `admin` calls which are typically not supported
	// by hosted node services. When not set, defaults to false.
	SkipGethAdminEnv = "SKIP_GETH_ADMIN"

	// MiddlewareVersion is the version of rosetta-ethereum.
	MiddlewareVersion = "0.0.4"

	// MainnetGethArguments are the arguments to start a mainnet geth instance.
	MainnetGethArguments = `--config=/app/ethereum/geth.toml --gcmode=archive --graphql`

	// GenesisBlockIndex is the index of the genesis block.
	GenesisBlockIndex = int64(0)
)

var (
	// RopstenGethArguments are the arguments to start a ropsten geth instance.
	RopstenGethArguments = fmt.Sprintf("%s --ropsten", MainnetGethArguments)

	// RinkebyGethArguments are the arguments to start a rinkeby geth instance.
	RinkebyGethArguments = fmt.Sprintf("%s --rinkeby", MainnetGethArguments)

	// GoerliGethArguments are the arguments to start a ropsten geth instance.
	GoerliGethArguments = fmt.Sprintf("%s --goerli", MainnetGethArguments)

	// MainnetGenesisBlockIdentifier is the *types.BlockIdentifier
	// of the mainnet genesis block.
	MainnetGenesisBlockIdentifier = &types.BlockIdentifier{
		Hash:  params.MainnetGenesisHash.Hex(),
		Index: GenesisBlockIndex,
	}

	// RopstenGenesisBlockIdentifier is the *types.BlockIdentifier
	// of the Ropsten genesis block.
	RopstenGenesisBlockIdentifier = &types.BlockIdentifier{
		Hash:  params.RopstenGenesisHash.Hex(),
		Index: GenesisBlockIndex,
	}

	// RinkebyGenesisBlockIdentifier is the *types.BlockIdentifier
	// of the Ropsten genesis block.
	RinkebyGenesisBlockIdentifier = &types.BlockIdentifier{
		Hash:  params.RinkebyGenesisHash.Hex(),
		Index: GenesisBlockIndex,
	}

	// GoerliGenesisBlockIdentifier is the *types.BlockIdentifier
	// of the Goerli genesis block.
	GoerliGenesisBlockIdentifier = &types.BlockIdentifier{
		Hash:  params.GoerliGenesisHash.Hex(),
		Index: GenesisBlockIndex,
	}
)

// LoadConfiguration attempts to create a new Configuration
// using the ENVs in the environment.
func LoadConfiguration() (*configuration.Configuration, error) {
	config := &configuration.Configuration{}

	mode := os.Getenv(ModeEnv)
	modeValue := configuration.Mode(mode)

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
			Blockchain: Blockchain,
			Network:    MainnetNetwork,
		}
		config.GenesisBlockIdentifier = MainnetGenesisBlockIdentifier
		config.ChainConfig = params.MainnetChainConfig
		config.GethArguments = MainnetGethArguments
	case Testnet, Ropsten, "":
		config.Network = &types.NetworkIdentifier{
			Blockchain: Blockchain,
			Network:    RopstenNetwork,
		}
		config.GenesisBlockIdentifier = RopstenGenesisBlockIdentifier
		config.ChainConfig = params.RopstenChainConfig
		config.GethArguments = RopstenGethArguments
	case Rinkeby:
		config.Network = &types.NetworkIdentifier{
			Blockchain: Blockchain,
			Network:    RinkebyNetwork,
		}
		config.GenesisBlockIdentifier = RinkebyGenesisBlockIdentifier
		config.ChainConfig = params.RinkebyChainConfig
		config.GethArguments = RinkebyGethArguments
	case Goerli:
		config.Network = &types.NetworkIdentifier{
			Blockchain: Blockchain,
			Network:    GoerliNetwork,
		}
		config.GenesisBlockIdentifier = GoerliGenesisBlockIdentifier
		config.ChainConfig = params.GoerliChainConfig
		config.GethArguments = GoerliGethArguments
	// case "":
	// 	return nil, errors.New("NETWORK must be populated")
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

	portValue := os.Getenv(PortEnv)
	if len(portValue) == 0 {
		return nil, errors.New("PORT must be populated")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || len(portValue) == 0 || port <= 0 {
		return nil, fmt.Errorf("%w: unable to parse port %s", err, portValue)
	}
	config.Port = port

	tokenFilter := os.Getenv(TokenFilterEnv)
	tokenFilterValue, err := strconv.ParseBool(tokenFilter)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to parse token filter %t", err, tokenFilterValue)
	}

	payload := []configuration.Token{}
	config.RosettaCfg = configuration.RosettaConfig{
		SupportRewardTx: true,
		TraceType:       configuration.GethNativeTrace,
		Currency: &RosettaTypes.Currency{
			Symbol:   "ETH",
			Decimals: 18,
		},
		TracePrefix:    "",
		FilterTokens:   tokenFilterValue,
		TokenWhiteList: payload,
	}

	return config, nil
}
