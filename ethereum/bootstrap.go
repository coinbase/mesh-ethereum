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

package ethereum

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/coinbase/rosetta-sdk-go/storage/modules"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/coinbase/rosetta-sdk-go/utils"
)

type genesis struct {
	Alloc map[string]genesisAllocation `json:"alloc"`
}

type genesisAllocation struct {
	Balance string `json:"balance"`
}

// GenerateBootstrapFile creates the bootstrap balances file
// for a particular genesis file.
func GenerateBootstrapFile(genesisFile string, outputFile string) error {
	var genesisAllocations genesis
	if err := utils.LoadAndParse(genesisFile, &genesisAllocations); err != nil {
		return fmt.Errorf("%w: could not load genesis file", err)
	}

	// Sort keys for deterministic genesis creation
	keys := make([]string, 0)
	formattedAllocations := map[string]string{}
	for k := range genesisAllocations.Alloc {
		checkAddr, ok := ChecksumAddress(k)
		if !ok {
			return fmt.Errorf("invalid address 0x%s", k)
		}
		keys = append(keys, checkAddr)
		formattedAllocations[checkAddr] = genesisAllocations.Alloc[k].Balance
	}
	sort.Strings(keys)

	// Write to file
	balances := []*modules.BootstrapBalance{}
	for _, k := range keys {
		v := formattedAllocations[k]
		bal, ok := new(big.Int).SetString(v[2:], 16)
		if !ok {
			return fmt.Errorf("cannot parse %s for integer", v)
		}

		if bal.Sign() == 0 {
			continue
		}

		balances = append(balances, &modules.BootstrapBalance{
			Account: &types.AccountIdentifier{
				Address: k,
			},
			Value:    bal.String(),
			Currency: Currency,
		})
	}

	if err := utils.SerializeAndWrite(outputFile, balances); err != nil {
		return fmt.Errorf("%w: could not write bootstrap balances", err)
	}

	return nil
}
