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

package client

import (
	"context"
	"io/ioutil"
	"math/big"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	evmClient "github.com/coinbase/rosetta-geth-sdk/client"
	mocks "github.com/coinbase/rosetta-geth-sdk/mocks/client"

	"github.com/ethereum/go-ethereum/common"
	EthTypes "github.com/ethereum/go-ethereum/core/types"
)

func TestCall_GetTransactionReceipt(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	ctx := context.Background()

	txHash := common.HexToHash("0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d")
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionReceipt",
		&txHash,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(**EthTypes.Receipt)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0x0046a7c3ca126864a3e851235ca6bf030300f9138f035f5f190e59ff9a4b22ff.json",
			)
			assert.NoError(t, err)

			*r = new(EthTypes.Receipt)

			assert.NoError(t, (*r).UnmarshalJSON(file))
		},
	).Once()

	rpcClient := &evmClient.RPCClient{
		JSONRPC: mockJSONRPC,
	}
	sdkClient := &evmClient.SDKClient{
		RPCClient: rpcClient,
	}

	c := &EthereumClient{
		*sdkClient,
	}

	h := common.HexToHash("0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d")
	gasPrice := big.NewInt(10000)
	myTx := EthTypes.NewTransaction(
		0,
		common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"),
		big.NewInt(0),
		0,
		gasPrice,
		nil,
	)

	loadedTxn := &evmClient.LoadedTransaction{
		TxHash:      &h,
		Transaction: myTx,
	}
	receipt, err := c.GetTransactionReceipt(ctx, loadedTxn)

	assert.NoError(t, err)
	assert.Equal(t, receipt.GasPrice, gasPrice)
	assert.Equal(t, receipt.GasUsed.String(), "98571")

	mockJSONRPC.AssertExpectations(t)
}
