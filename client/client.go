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
	"fmt"
	"github.com/coinbase/rosetta-geth-sdk/services"
	RosettaTypes "github.com/coinbase/rosetta-sdk-go/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"log"
	"math/big"

	evmClient "github.com/coinbase/rosetta-geth-sdk/client"
	"github.com/coinbase/rosetta-geth-sdk/examples/ethereum/config"
	sdkTypes "github.com/coinbase/rosetta-geth-sdk/types"
	EthTypes "github.com/ethereum/go-ethereum/core/types"
)

type EthereumClient struct {
	// Use embedding for inheritance. So all the methods of the SDKClient
	// are instantly available on EthereumClient.
	evmClient.SDKClient
}

func (c *EthereumClient) ParseOps(
	tx *evmClient.LoadedTransaction,
) ([]*RosettaTypes.Operation, error) {
	var ops []*RosettaTypes.Operation

	// Compute fee operations
	feeOps := services.FeeOps(tx)
	ops = append(ops, feeOps...)

	traceOps := services.TraceOps(tx.Trace, len(ops))
	ops = append(ops, traceOps...)

	return ops, nil
}

func (c *EthereumClient) GetBlockReceipts(
	ctx context.Context,
	blockHash common.Hash,
	txs []evmClient.RPCTransaction,
	baseFee *big.Int,
) ([]*evmClient.RosettaTxReceipt, error) {
	receipts := make([]*evmClient.RosettaTxReceipt, len(txs))
	if len(txs) == 0 {
		return receipts, nil
	}

	ethReceipts := make([]*EthTypes.Receipt, len(txs))
	reqs := make([]rpc.BatchElem, len(txs))
	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{txs[i].TxExtraInfo.TxHash.String()},
			Result: &ethReceipts[i],
		}
	}
	if err := c.BatchCallContext(ctx, reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, reqs[i].Error
		}

		gasPrice, err := evmClient.EffectiveGasPrice(txs[i].Tx, baseFee)
		if err != nil {
			return nil, err
		}
		gasUsed := new(big.Int).SetUint64(ethReceipts[i].GasUsed)
		feeAmount := new(big.Int).Mul(gasUsed, gasPrice)

		receipt := &evmClient.RosettaTxReceipt{
			GasPrice:       gasPrice,
			GasUsed:        gasUsed,
			Logs:           ethReceipts[i].Logs,
			RawMessage:     nil,
			TransactionFee: feeAmount,
		}

		receipts[i] = receipt

		if ethReceipts[i] == nil {
			return nil, fmt.Errorf("got empty receipt for %x", txs[i].Tx.Hash().Hex())
		}

		if ethReceipts[i].BlockHash != blockHash {
			return nil, fmt.Errorf(
				"%w: expected block hash %s for Transaction but got %s",
				sdkTypes.ErrClientBlockOrphaned,
				blockHash.Hex(),
				ethReceipts[i].BlockHash.Hex(),
			)
		}
	}

	return receipts, nil
}

func (c *EthereumClient) GetTransactionReceipt(
	ctx context.Context,
	tx *evmClient.LoadedTransaction,
) (*evmClient.RosettaTxReceipt, error) {
	var r *EthTypes.Receipt
	err := c.CallContext(ctx, &r, "eth_getTransactionReceipt", tx.TxHash)
	if err == nil {
		if r == nil {
			return nil, ethereum.NotFound
		}
	}
	gasPrice, err := evmClient.EffectiveGasPrice(tx.Transaction, tx.BaseFee)
	if err != nil {
		return nil, err
	}
	gasUsed := new(big.Int).SetUint64(r.GasUsed)
	feeAmount := new(big.Int).Mul(gasUsed, gasPrice)

	return &evmClient.RosettaTxReceipt{
		GasPrice:       gasPrice,
		GasUsed:        gasUsed,
		Logs:           r.Logs,
		RawMessage:     nil,
		TransactionFee: feeAmount,
	}, err
}

// GetNativeTransferGasLimit is Ethereum's custom implementation of estimating gas.
func (c *EthereumClient) GetNativeTransferGasLimit(ctx context.Context, toAddress string,
	fromAddress string, value *big.Int) (uint64, error) {
	if len(toAddress) == 0 || value == nil {
		// We guard against malformed inputs that may have been generated using
		// a previous version of asset's rosetta
		return 21000, nil
	}
	to := common.HexToAddress(toAddress)
	return c.EstimateGas(ctx, ethereum.CallMsg{
		From:  common.HexToAddress(fromAddress),
		To:    &to,
		Value: big.NewInt(0),
	})
}

// NewEthereumClient creates a eth client that can interact with
// Ethereum network.
func NewEthereumClient() (*EthereumClient, error) {
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatalln("%w: unable to load configuration", err)
	}

	// Use SDK to quickly create a client that support JSON RPC calls
	evmClient, err := evmClient.NewClient(cfg, nil)

	if err != nil {
		log.Fatalln("%w: cannot initialize client", err)
		return nil, err
	}

	// Use embedding for inheritance. So all the methods of the SDKClient
	// are instantly available on EthereumClient.
	p := &EthereumClient{
		*evmClient,
	}

	return p, err
}
