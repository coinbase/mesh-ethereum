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

package services

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// Client is used by the services to get block
// data and to submit transactions.
type Client interface {
	Status(context.Context) (
		*types.BlockIdentifier,
		int64,
		*types.SyncStatus,
		[]*types.Peer,
		error,
	)

	Block(
		context.Context,
		*types.PartialBlockIdentifier,
	) (*types.Block, error)

	Transaction(
		context.Context,
		*types.BlockIdentifier,
		*types.TransactionIdentifier,
	) (*types.Transaction, error)

	Balance(
		context.Context,
		*types.AccountIdentifier,
		*types.PartialBlockIdentifier,
	) (*types.AccountBalanceResponse, error)

	PendingNonceAt(context.Context, common.Address) (uint64, error)

	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	SendTransaction(ctx context.Context, tx *ethTypes.Transaction) error

	GetMempool(ctx context.Context) (*types.MempoolResponse, error)

	Call(
		ctx context.Context,
		request *types.CallRequest,
	) (*types.CallResponse, error)
}

type options struct {
	From string `json:"from"`
}

type metadata struct {
	Nonce    uint64   `json:"nonce"`
	GasPrice *big.Int `json:"gas_price"`
}

type metadataWire struct {
	Nonce    string `json:"nonce"`
	GasPrice string `json:"gas_price"`
}

func (m *metadata) MarshalJSON() ([]byte, error) {
	mw := &metadataWire{
		Nonce:    hexutil.Uint64(m.Nonce).String(),
		GasPrice: hexutil.EncodeBig(m.GasPrice),
	}

	return json.Marshal(mw)
}

func (m *metadata) UnmarshalJSON(data []byte) error {
	var mw metadataWire
	if err := json.Unmarshal(data, &mw); err != nil {
		return err
	}

	nonce, err := hexutil.DecodeUint64(mw.Nonce)
	if err != nil {
		return err
	}

	gasPrice, err := hexutil.DecodeBig(mw.GasPrice)
	if err != nil {
		return err
	}

	m.GasPrice = gasPrice
	m.Nonce = nonce
	return nil
}

type parseMetadata struct {
	Nonce    uint64   `json:"nonce"`
	GasPrice *big.Int `json:"gas_price"`
	ChainID  *big.Int `json:"chain_id"`
}

type parseMetadataWire struct {
	Nonce    string `json:"nonce"`
	GasPrice string `json:"gas_price"`
	ChainID  string `json:"chain_id"`
}

func (p *parseMetadata) MarshalJSON() ([]byte, error) {
	pmw := &parseMetadataWire{
		Nonce:    hexutil.Uint64(p.Nonce).String(),
		GasPrice: hexutil.EncodeBig(p.GasPrice),
		ChainID:  hexutil.EncodeBig(p.ChainID),
	}

	return json.Marshal(pmw)
}

type transaction struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Value    *big.Int `json:"value"`
	Data     []byte   `json:"data"`
	Nonce    uint64   `json:"nonce"`
	GasPrice *big.Int `json:"gas_price"`
	GasLimit uint64   `json:"gas"`
	ChainID  *big.Int `json:"chain_id"`
}

type transactionWire struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	Nonce    string `json:"nonce"`
	GasPrice string `json:"gas_price"`
	GasLimit string `json:"gas"`
	ChainID  string `json:"chain_id"`
}

func (t *transaction) MarshalJSON() ([]byte, error) {
	tw := &transactionWire{
		From:     t.From,
		To:       t.To,
		Value:    hexutil.EncodeBig(t.Value),
		Data:     hexutil.Encode(t.Data),
		Nonce:    hexutil.EncodeUint64(t.Nonce),
		GasPrice: hexutil.EncodeBig(t.GasPrice),
		GasLimit: hexutil.EncodeUint64(t.GasLimit),
		ChainID:  hexutil.EncodeBig(t.ChainID),
	}

	return json.Marshal(tw)
}

func (t *transaction) UnmarshalJSON(data []byte) error {
	var tw transactionWire
	if err := json.Unmarshal(data, &tw); err != nil {
		return err
	}

	value, err := hexutil.DecodeBig(tw.Value)
	if err != nil {
		return err
	}

	twData, err := hexutil.Decode(tw.Data)
	if err != nil {
		return err
	}

	nonce, err := hexutil.DecodeUint64(tw.Nonce)
	if err != nil {
		return err
	}

	gasPrice, err := hexutil.DecodeBig(tw.GasPrice)
	if err != nil {
		return err
	}

	gasLimit, err := hexutil.DecodeUint64(tw.GasLimit)
	if err != nil {
		return err
	}

	chainID, err := hexutil.DecodeBig(tw.ChainID)
	if err != nil {
		return err
	}

	t.From = tw.From
	t.To = tw.To
	t.Value = value
	t.Data = twData
	t.Nonce = nonce
	t.GasPrice = gasPrice
	t.GasLimit = gasLimit
	t.ChainID = chainID
	t.GasPrice = gasPrice
	return nil
}
