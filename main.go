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

package main

import (
	"log"

	"github.com/coinbase/rosetta-geth-sdk/examples/ethereum/client"
	"github.com/coinbase/rosetta-geth-sdk/examples/ethereum/config"
	sdkTypes "github.com/coinbase/rosetta-geth-sdk/types"
	"github.com/coinbase/rosetta-geth-sdk/utils"
)

func main() {
	// Load configuration using the ENVs in the environment.
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalln("%w: unable to load configuration", err)
	}

	// Load all the supported operation types, status
	types := sdkTypes.LoadTypes()
	errors := sdkTypes.Errors

	// Create a new ethereum client by leveraging SDK functionalities
	client, err := client.NewEthereumClient()
	if err != nil {
		log.Fatalln("%w: cannot initialize client", err)
	}

	// Bootstrap to start the Rosetta API server
	err = utils.BootStrap(cfg, types, errors, client)
	if err != nil {
		log.Fatalln("%w: unable to bootstrap Rosetta server", err)
	}
}
