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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	RosettaTypes "github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/types"
	EthTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/semaphore"
)

const (
	gethHTTPTimeout = 120 * time.Second

	maxTraceConcurrency  = int64(16) // nolint:gomnd
	semaphoreTraceWeight = int64(1)  // nolint:gomnd
)

// Client allows for querying a set of specific Ethereum endpoints in an
// idempotent manner. Client relies on the eth_*, debug_*, and admin_*
// methods and on the graphql endpoint.
//
// Client borrows HEAVILY from https://github.com/ethereum/go-ethereum/tree/master/ethclient.
type Client struct {
	p  *params.ChainConfig
	tc *eth.TraceConfig

	c JSONRPC
	g GraphQL

	traceSemaphore *semaphore.Weighted

	skipAdminCalls bool
}

// NewClient creates a Client that from the provided url and params.
func NewClient(url string, params *params.ChainConfig, skipAdminCalls bool, headers []*HTTPHeader) (*Client, error) {
	c, err := rpc.DialHTTPWithClient(url, &http.Client{
		Timeout: gethHTTPTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: unable to dial node", err)
	}

	for _, header := range headers {
		c.SetHeader(header.Key, header.Value)
	}

	tc, err := loadTraceConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: unable to load trace config", err)
	}

	g, err := newGraphQLClient(url, headers)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to create GraphQL client", err)
	}

	return &Client{params, tc, c, g, semaphore.NewWeighted(maxTraceConcurrency), skipAdminCalls}, nil
}

// Close shuts down the RPC client connection.
func (ec *Client) Close() {
	ec.c.Close()
}

// Status returns geth status information
// for determining node healthiness.
func (ec *Client) Status(ctx context.Context) (
	*RosettaTypes.BlockIdentifier,
	int64,
	*RosettaTypes.SyncStatus,
	[]*RosettaTypes.Peer,
	error,
) {
	header, err := ec.blockHeader(ctx, nil)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	progress, err := ec.syncProgress(ctx)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	var syncStatus *RosettaTypes.SyncStatus
	if progress != nil {
		currentIndex := int64(progress.CurrentBlock)
		targetIndex := int64(progress.HighestBlock)

		syncStatus = &RosettaTypes.SyncStatus{
			CurrentIndex: &currentIndex,
			TargetIndex:  &targetIndex,
		}
	}

	peers, err := ec.peers(ctx)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	return &RosettaTypes.BlockIdentifier{
			Hash:  header.Hash().Hex(),
			Index: header.Number.Int64(),
		},
		convertTime(header.Time),
		syncStatus,
		peers,
		nil
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (ec *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_getTransactionCount", account, "pending")
	return uint64(result), err
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (ec *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// Peers retrieves all peers of the node.
func (ec *Client) peers(ctx context.Context) ([]*RosettaTypes.Peer, error) {
	var info []*p2p.PeerInfo

	if ec.skipAdminCalls {
		return []*RosettaTypes.Peer{}, nil
	}

	if err := ec.c.CallContext(ctx, &info, "admin_peers"); err != nil {
		return nil, err
	}

	peers := make([]*RosettaTypes.Peer, len(info))
	for i, peerInfo := range info {
		peers[i] = &RosettaTypes.Peer{
			PeerID: peerInfo.ID,
			Metadata: map[string]interface{}{
				"name":      peerInfo.Name,
				"enode":     peerInfo.Enode,
				"caps":      peerInfo.Caps,
				"enr":       peerInfo.ENR,
				"protocols": peerInfo.Protocols,
			},
		}
	}

	return peers, nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.c.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(data))
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}

// Block returns a populated block at the *RosettaTypes.PartialBlockIdentifier.
// If neither the hash or index is populated in the *RosettaTypes.PartialBlockIdentifier,
// the current block is returned.
func (ec *Client) Block(
	ctx context.Context,
	blockIdentifier *RosettaTypes.PartialBlockIdentifier,
) (*RosettaTypes.Block, error) {
	if blockIdentifier != nil {
		if blockIdentifier.Hash != nil {
			return ec.getParsedBlock(ctx, "eth_getBlockByHash", *blockIdentifier.Hash, true)
		}

		if blockIdentifier.Index != nil {
			return ec.getParsedBlock(
				ctx,
				"eth_getBlockByNumber",
				toBlockNumArg(big.NewInt(*blockIdentifier.Index)),
				true,
			)
		}
	}

	return ec.getParsedBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(nil), true)
}

// Header returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ec *Client) blockHeader(ctx context.Context, number *big.Int) (*types.Header, error) {
	var head *types.Header
	err := ec.c.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		return nil, ethereum.NotFound
	}

	return head, err
}

type rpcBlock struct {
	Hash         common.Hash      `json:"hash"`
	Transactions []rpcTransaction `json:"transactions"`
	UncleHashes  []common.Hash    `json:"uncles"`
}

func (ec *Client) getUncles(
	ctx context.Context,
	head *types.Header,
	body *rpcBlock,
) ([]*types.Header, error) {
	// Quick-verify transaction and uncle lists. This mostly helps with debugging the server.
	if head.UncleHash == types.EmptyUncleHash && len(body.UncleHashes) > 0 {
		return nil, fmt.Errorf(
			"server returned non-empty uncle list but block header indicates no uncles",
		)
	}
	if head.UncleHash != types.EmptyUncleHash && len(body.UncleHashes) == 0 {
		return nil, fmt.Errorf(
			"server returned empty uncle list but block header indicates uncles",
		)
	}
	if head.TxHash == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, fmt.Errorf(
			"server returned non-empty transaction list but block header indicates no transactions",
		)
	}
	if head.TxHash != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, fmt.Errorf(
			"server returned empty transaction list but block header indicates transactions",
		)
	}
	// Load uncles because they are not included in the block response.
	var uncles []*types.Header
	if len(body.UncleHashes) > 0 {
		uncles = make([]*types.Header, len(body.UncleHashes))
		reqs := make([]rpc.BatchElem, len(body.UncleHashes))
		for i := range reqs {
			reqs[i] = rpc.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))},
				Result: &uncles[i],
			}
		}
		if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
			if uncles[i] == nil {
				return nil, fmt.Errorf(
					"got null header for uncle %d of block %x",
					i,
					body.Hash[:],
				)
			}
		}
	}

	return uncles, nil
}

func (ec *Client) getBlock(
	ctx context.Context,
	blockMethod string,
	args ...interface{},
) (
	*types.Block,
	[]*loadedTransaction,
	error,
) {
	var raw json.RawMessage
	err := ec.c.CallContext(ctx, &raw, blockMethod, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: block fetch failed", err)
	} else if len(raw) == 0 {
		return nil, nil, ethereum.NotFound
	}

	// Decode header and transactions
	var head types.Header
	var body rpcBlock
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, nil, err
	}

	uncles, err := ec.getUncles(ctx, &head, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: unable to get uncles", err)
	}

	// Get all transaction receipts
	receipts, err := ec.getBlockReceipts(ctx, body.Hash, body.Transactions)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: could not get receipts for %x", err, body.Hash[:])
	}

	// Get block traces (not possible to make idempotent block transaction trace requests)
	//
	// We fetch traces last because we want to avoid limiting the number of other
	// block-related data fetches we perform concurrently (we limit the number of
	// concurrent traces that are computed to 16 to avoid overwhelming geth).
	var traces []*rpcCall
	var rawTraces []*rpcRawCall
	var addTraces bool
	if head.Number.Int64() != GenesisBlockIndex { // not possible to get traces at genesis
		addTraces = true
		traces, rawTraces, err = ec.getBlockTraces(ctx, body.Hash)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: could not get traces for %x", err, body.Hash[:])
		}
	}

	// Convert all txs to loaded txs
	txs := make([]*types.Transaction, len(body.Transactions))
	loadedTxs := make([]*loadedTransaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		txs[i] = tx.tx
		receipt := receipts[i]
		gasUsedBig := new(big.Int).SetUint64(receipt.GasUsed)
		feeAmount := gasUsedBig.Mul(gasUsedBig, txs[i].GasPrice())

		loadedTxs[i] = tx.LoadedTransaction()
		loadedTxs[i].Transaction = txs[i]
		loadedTxs[i].FeeAmount = feeAmount
		loadedTxs[i].Miner = MustChecksum(head.Coinbase.Hex())
		loadedTxs[i].Receipt = receipt

		// Continue if calls does not exist (occurs at genesis)
		if !addTraces {
			continue
		}

		loadedTxs[i].Trace = traces[i].Result
		loadedTxs[i].RawTrace = rawTraces[i].Result
	}

	return types.NewBlockWithHeader(&head).WithBody(txs, uncles), loadedTxs, nil
}

func (ec *Client) getBlockTraces(
	ctx context.Context,
	blockHash common.Hash,
) ([]*rpcCall, []*rpcRawCall, error) {
	if err := ec.traceSemaphore.Acquire(ctx, semaphoreTraceWeight); err != nil {
		return nil, nil, err
	}
	defer ec.traceSemaphore.Release(semaphoreTraceWeight)

	var calls []*rpcCall
	var rawCalls []*rpcRawCall
	var raw json.RawMessage
	err := ec.c.CallContext(ctx, &raw, "debug_traceBlockByHash", blockHash, ec.tc)
	if err != nil {
		return nil, nil, err
	}

	// Decode []*rpcCall
	if err := json.Unmarshal(raw, &calls); err != nil {
		return nil, nil, err
	}

	// Decode []*rpcRawCall
	if err := json.Unmarshal(raw, &rawCalls); err != nil {
		return nil, nil, err
	}

	return calls, rawCalls, nil
}

func (ec *Client) getBlockReceipts(
	ctx context.Context,
	blockHash common.Hash,
	txs []rpcTransaction,
) ([]*types.Receipt, error) {
	receipts := make([]*types.Receipt, len(txs))
	if len(txs) == 0 {
		return receipts, nil
	}

	reqs := make([]rpc.BatchElem, len(txs))
	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{txs[i].tx.Hash().Hex()},
			Result: &receipts[i],
		}
	}
	if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, reqs[i].Error
		}
		if receipts[i] == nil {
			return nil, fmt.Errorf("got empty receipt for %x", txs[i].tx.Hash().Hex())
		}

		if receipts[i].BlockHash != blockHash {
			return nil, fmt.Errorf(
				"%w: expected block hash %s for transaction but got %s",
				ErrBlockOrphaned,
				blockHash.Hex(),
				receipts[i].BlockHash.Hex(),
			)
		}
	}

	return receipts, nil
}

type rpcCall struct {
	Result *Call `json:"result"`
}

type rpcRawCall struct {
	Result json.RawMessage `json:"result"`
}

// Call is an Ethereum debug trace.
type Call struct {
	Type         string         `json:"type"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Value        *big.Int       `json:"value"`
	GasUsed      *big.Int       `json:"gasUsed"`
	Revert       bool
	ErrorMessage string  `json:"error"`
	Calls        []*Call `json:"calls"`
}

type flatCall struct {
	Type         string         `json:"type"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Value        *big.Int       `json:"value"`
	GasUsed      *big.Int       `json:"gasUsed"`
	Revert       bool
	ErrorMessage string `json:"error"`
}

func (t *Call) flatten() *flatCall {
	return &flatCall{
		Type:         t.Type,
		From:         t.From,
		To:           t.To,
		Value:        t.Value,
		GasUsed:      t.GasUsed,
		Revert:       t.Revert,
		ErrorMessage: t.ErrorMessage,
	}
}

// UnmarshalJSON is a custom unmarshaler for Call.
func (t *Call) UnmarshalJSON(input []byte) error {
	type CustomTrace struct {
		Type         string         `json:"type"`
		From         common.Address `json:"from"`
		To           common.Address `json:"to"`
		Value        *hexutil.Big   `json:"value"`
		GasUsed      *hexutil.Big   `json:"gasUsed"`
		Revert       bool
		ErrorMessage string  `json:"error"`
		Calls        []*Call `json:"calls"`
	}
	var dec CustomTrace
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	t.Type = dec.Type
	t.From = dec.From
	t.To = dec.To
	if dec.Value != nil {
		t.Value = (*big.Int)(dec.Value)
	} else {
		t.Value = new(big.Int)
	}
	if dec.GasUsed != nil {
		t.GasUsed = (*big.Int)(dec.Value)
	} else {
		t.GasUsed = new(big.Int)
	}
	if dec.ErrorMessage != "" {
		// Any error surfaced by the decoder means that the transaction
		// has reverted.
		t.Revert = true
	}
	t.ErrorMessage = dec.ErrorMessage
	t.Calls = dec.Calls
	return nil
}

// flattenTraces recursively flattens all traces.
func flattenTraces(data *Call, flattened []*flatCall) []*flatCall {
	results := append(flattened, data.flatten())
	for _, child := range data.Calls {
		// Ensure all children of a reverted call
		// are also reverted!
		if data.Revert {
			child.Revert = true

			// Copy error message from parent
			// if child does not have one
			if len(child.ErrorMessage) == 0 {
				child.ErrorMessage = data.ErrorMessage
			}
		}

		children := flattenTraces(child, flattened)
		results = append(results, children...)
	}
	return results
}

// traceOps returns all *RosettaTypes.Operation for a given
// array of flattened traces.
func traceOps(calls []*flatCall, startIndex int) []*RosettaTypes.Operation { // nolint: gocognit
	var ops []*RosettaTypes.Operation
	if len(calls) == 0 {
		return ops
	}

	destroyedAccounts := map[string]*big.Int{}
	for _, trace := range calls {
		// Handle partial transaction success
		metadata := map[string]interface{}{}
		opStatus := SuccessStatus
		if trace.Revert {
			opStatus = FailureStatus
			metadata["error"] = trace.ErrorMessage
		}

		var zeroValue bool
		if trace.Value.Sign() == 0 {
			zeroValue = true
		}

		// Skip all 0 value CallType operations (TODO: make optional to include)
		//
		// We can't continue here because we may need to adjust our destroyed
		// accounts map if a CallTYpe operation resurrects an account.
		shouldAdd := true
		if zeroValue && CallType(trace.Type) {
			shouldAdd = false
		}

		// Checksum addresses
		from := MustChecksum(trace.From.String())
		to := MustChecksum(trace.To.String())

		if shouldAdd {
			fromOp := &RosettaTypes.Operation{
				OperationIdentifier: &RosettaTypes.OperationIdentifier{
					Index: int64(len(ops) + startIndex),
				},
				Type:   trace.Type,
				Status: RosettaTypes.String(opStatus),
				Account: &RosettaTypes.AccountIdentifier{
					Address: from,
				},
				Amount: &RosettaTypes.Amount{
					Value:    new(big.Int).Neg(trace.Value).String(),
					Currency: Currency,
				},
				Metadata: metadata,
			}
			if zeroValue {
				fromOp.Amount = nil
			} else {
				_, destroyed := destroyedAccounts[from]
				if destroyed && opStatus == SuccessStatus {
					destroyedAccounts[from] = new(big.Int).Sub(destroyedAccounts[from], trace.Value)
				}
			}

			ops = append(ops, fromOp)
		}

		// Add to destroyed accounts if SELFDESTRUCT
		// and overwrite existing balance.
		if trace.Type == SelfDestructOpType {
			destroyedAccounts[from] = new(big.Int)

			// If destination of of SELFDESTRUCT is self,
			// we should skip. In the EVM, the balance is reset
			// after the balance is increased on the destination
			// so this is a no-op.
			if from == to {
				continue
			}
		}

		// Skip empty to addresses (this may not
		// actually occur but leaving it as a
		// sanity check)
		if len(trace.To.String()) == 0 {
			continue
		}

		// If the account is resurrected, we remove it from
		// the destroyed accounts map.
		if CreateType(trace.Type) {
			delete(destroyedAccounts, to)
		}

		if shouldAdd {
			lastOpIndex := ops[len(ops)-1].OperationIdentifier.Index
			toOp := &RosettaTypes.Operation{
				OperationIdentifier: &RosettaTypes.OperationIdentifier{
					Index: lastOpIndex + 1,
				},
				RelatedOperations: []*RosettaTypes.OperationIdentifier{
					{
						Index: lastOpIndex,
					},
				},
				Type:   trace.Type,
				Status: RosettaTypes.String(opStatus),
				Account: &RosettaTypes.AccountIdentifier{
					Address: to,
				},
				Amount: &RosettaTypes.Amount{
					Value:    trace.Value.String(),
					Currency: Currency,
				},
				Metadata: metadata,
			}
			if zeroValue {
				toOp.Amount = nil
			} else {
				_, destroyed := destroyedAccounts[to]
				if destroyed && opStatus == SuccessStatus {
					destroyedAccounts[to] = new(big.Int).Add(destroyedAccounts[to], trace.Value)
				}
			}

			ops = append(ops, toOp)
		}
	}

	// Zero-out all destroyed accounts that are removed
	// during transaction finalization.
	for acct, val := range destroyedAccounts {
		if val.Sign() == 0 {
			continue
		}

		if val.Sign() < 0 {
			log.Fatalf("negative balance for suicided account %s: %s\n", acct, val.String())
		}

		ops = append(ops, &RosettaTypes.Operation{
			OperationIdentifier: &RosettaTypes.OperationIdentifier{
				Index: ops[len(ops)-1].OperationIdentifier.Index + 1,
			},
			Type:   DestructOpType,
			Status: RosettaTypes.String(SuccessStatus),
			Account: &RosettaTypes.AccountIdentifier{
				Address: acct,
			},
			Amount: &RosettaTypes.Amount{
				Value:    new(big.Int).Neg(val).String(),
				Currency: Currency,
			},
		})
	}

	return ops
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func (tx *rpcTransaction) LoadedTransaction() *loadedTransaction {
	ethTx := &loadedTransaction{
		Transaction: tx.tx,
		From:        tx.txExtraInfo.From,
		BlockNumber: tx.txExtraInfo.BlockNumber,
		BlockHash:   tx.txExtraInfo.BlockHash,
	}
	return ethTx
}

type loadedTransaction struct {
	Transaction *types.Transaction
	From        *common.Address
	BlockNumber *string
	BlockHash   *common.Hash
	FeeAmount   *big.Int
	Miner       string
	Status      bool

	Trace    *Call
	RawTrace json.RawMessage
	Receipt  *types.Receipt
}

func feeOps(tx *loadedTransaction) []*RosettaTypes.Operation {
	return []*RosettaTypes.Operation{
		{
			OperationIdentifier: &RosettaTypes.OperationIdentifier{
				Index: 0,
			},
			Type:   FeeOpType,
			Status: RosettaTypes.String(SuccessStatus),
			Account: &RosettaTypes.AccountIdentifier{
				Address: MustChecksum(tx.From.String()),
			},
			Amount: &RosettaTypes.Amount{
				Value:    new(big.Int).Neg(tx.FeeAmount).String(),
				Currency: Currency,
			},
		},

		{
			OperationIdentifier: &RosettaTypes.OperationIdentifier{
				Index: 1,
			},
			RelatedOperations: []*RosettaTypes.OperationIdentifier{
				{
					Index: 0,
				},
			},
			Type:   FeeOpType,
			Status: RosettaTypes.String(SuccessStatus),
			Account: &RosettaTypes.AccountIdentifier{
				Address: MustChecksum(tx.Miner),
			},
			Amount: &RosettaTypes.Amount{
				Value:    tx.FeeAmount.String(),
				Currency: Currency,
			},
		},
	}
}

// transactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) transactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	var r *types.Receipt
	err := ec.c.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash)
	if err == nil {
		if r == nil {
			return nil, ethereum.NotFound
		}
	}

	return r, err
}

func (ec *Client) getParsedBlock(
	ctx context.Context,
	blockMethod string,
	args ...interface{},
) (
	*RosettaTypes.Block,
	error,
) {
	block, loadedTransactions, err := ec.getBlock(ctx, blockMethod, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: could not get block", err)
	}

	blockIdentifier := &RosettaTypes.BlockIdentifier{
		Hash:  block.Hash().String(),
		Index: block.Number().Int64(),
	}

	parentBlockIdentifier := blockIdentifier
	if blockIdentifier.Index != GenesisBlockIndex {
		parentBlockIdentifier = &RosettaTypes.BlockIdentifier{
			Hash:  block.ParentHash().Hex(),
			Index: blockIdentifier.Index - 1,
		}
	}

	txs, err := ec.populateTransactions(blockIdentifier, block, loadedTransactions)
	if err != nil {
		return nil, err
	}

	return &RosettaTypes.Block{
		BlockIdentifier:       blockIdentifier,
		ParentBlockIdentifier: parentBlockIdentifier,
		Timestamp:             convertTime(block.Time()),
		Transactions:          txs,
	}, nil
}

func convertTime(time uint64) int64 {
	return int64(time) * 1000
}

func (ec *Client) populateTransactions(
	blockIdentifier *RosettaTypes.BlockIdentifier,
	block *EthTypes.Block,
	loadedTransactions []*loadedTransaction,
) ([]*RosettaTypes.Transaction, error) {
	transactions := make(
		[]*RosettaTypes.Transaction,
		len(block.Transactions())+1, // include reward tx
	)

	// Compute reward transaction (block + uncle reward)
	transactions[0] = ec.blockRewardTransaction(
		blockIdentifier,
		block.Coinbase().String(),
		block.Uncles(),
	)

	for i, tx := range loadedTransactions {
		transaction, err := ec.populateTransaction(
			tx,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: cannot parse %s", err, tx.Transaction.Hash().Hex())
		}

		transactions[i+1] = transaction
	}

	return transactions, nil
}

func (ec *Client) populateTransaction(
	tx *loadedTransaction,
) (*RosettaTypes.Transaction, error) {
	ops := []*RosettaTypes.Operation{}

	// Compute fee operations
	feeOps := feeOps(tx)
	ops = append(ops, feeOps...)

	// Compute trace operations
	traces := flattenTraces(tx.Trace, []*flatCall{})

	traceOps := traceOps(traces, len(ops))
	ops = append(ops, traceOps...)

	// Marshal receipt and trace data
	// TODO: replace with marshalJSONMap (used in `services`)
	receiptBytes, err := tx.Receipt.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var receiptMap map[string]interface{}
	if err := json.Unmarshal(receiptBytes, &receiptMap); err != nil {
		return nil, err
	}

	var traceMap map[string]interface{}
	if err := json.Unmarshal(tx.RawTrace, &traceMap); err != nil {
		return nil, err
	}

	populatedTransaction := &RosettaTypes.Transaction{
		TransactionIdentifier: &RosettaTypes.TransactionIdentifier{
			Hash: tx.Transaction.Hash().Hex(),
		},
		Operations: ops,
		Metadata: map[string]interface{}{
			"gas_limit": hexutil.EncodeUint64(tx.Transaction.Gas()),
			"gas_price": hexutil.EncodeBig(tx.Transaction.GasPrice()),
			"receipt":   receiptMap,
			"trace":     traceMap,
		},
	}

	return populatedTransaction, nil
}

// miningReward returns the mining reward
// for a given block height.
//
// Source:
// https://github.com/ethereum/go-ethereum/master/consensus/ethash/consensus.go#L619-L645
func (ec *Client) miningReward(
	currentBlock *big.Int,
) int64 {
	if currentBlock.Int64() == int64(0) {
		return big.NewInt(0).Int64()
	}

	blockReward := ethash.FrontierBlockReward.Int64()
	if ec.p.IsByzantium(currentBlock) {
		blockReward = ethash.ByzantiumBlockReward.Int64()
	}

	if ec.p.IsConstantinople(currentBlock) {
		blockReward = ethash.ConstantinopleBlockReward.Int64()
	}

	return blockReward
}

func (ec *Client) blockRewardTransaction(
	blockIdentifier *RosettaTypes.BlockIdentifier,
	miner string,
	uncles []*EthTypes.Header,
) *RosettaTypes.Transaction {
	var ops []*RosettaTypes.Operation
	miningReward := ec.miningReward(big.NewInt(blockIdentifier.Index))

	// Calculate miner rewards
	minerReward := miningReward
	numUncles := len(uncles)
	if len(uncles) > 0 {
		reward := new(big.Float)
		uncleReward := float64(numUncles) / UnclesRewardMultiplier
		rewardFloat := reward.Mul(big.NewFloat(uncleReward), big.NewFloat(float64(miningReward)))
		rewardInt, _ := rewardFloat.Int64()
		minerReward += rewardInt
	}

	miningRewardOp := &RosettaTypes.Operation{
		OperationIdentifier: &RosettaTypes.OperationIdentifier{
			Index: 0,
		},
		Type:   MinerRewardOpType,
		Status: RosettaTypes.String(SuccessStatus),
		Account: &RosettaTypes.AccountIdentifier{
			Address: MustChecksum(miner),
		},
		Amount: &RosettaTypes.Amount{
			Value:    strconv.FormatInt(minerReward, 10),
			Currency: Currency,
		},
	}
	ops = append(ops, miningRewardOp)

	// Calculate uncle rewards
	for _, b := range uncles {
		uncleMiner := b.Coinbase.String()
		uncleBlock := b.Number.Int64()
		uncleRewardBlock := new(
			big.Int,
		).Mul(
			big.NewInt(uncleBlock+MaxUncleDepth-blockIdentifier.Index),
			big.NewInt(miningReward/MaxUncleDepth),
		)

		uncleRewardOp := &RosettaTypes.Operation{
			OperationIdentifier: &RosettaTypes.OperationIdentifier{
				Index: int64(len(ops)),
			},
			Type:   UncleRewardOpType,
			Status: RosettaTypes.String(SuccessStatus),
			Account: &RosettaTypes.AccountIdentifier{
				Address: MustChecksum(uncleMiner),
			},
			Amount: &RosettaTypes.Amount{
				Value:    uncleRewardBlock.String(),
				Currency: Currency,
			},
		}
		ops = append(ops, uncleRewardOp)
	}

	return &RosettaTypes.Transaction{
		TransactionIdentifier: &RosettaTypes.TransactionIdentifier{
			Hash: blockIdentifier.Hash,
		},
		Operations: ops,
	}
}

type rpcProgress struct {
	StartingBlock hexutil.Uint64
	CurrentBlock  hexutil.Uint64
	HighestBlock  hexutil.Uint64
	PulledStates  hexutil.Uint64
	KnownStates   hexutil.Uint64
}

// syncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (ec *Client) syncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	var raw json.RawMessage
	if err := ec.c.CallContext(ctx, &raw, "eth_syncing"); err != nil {
		return nil, err
	}

	var syncing bool
	if err := json.Unmarshal(raw, &syncing); err == nil {
		return nil, nil // Not syncing (always false)
	}

	var progress rpcProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, err
	}

	return &ethereum.SyncProgress{
		StartingBlock: uint64(progress.StartingBlock),
		CurrentBlock:  uint64(progress.CurrentBlock),
		HighestBlock:  uint64(progress.HighestBlock),
		PulledStates:  uint64(progress.PulledStates),
		KnownStates:   uint64(progress.KnownStates),
	}, nil
}

type graphqlBalance struct {
	Errors []struct {
		Message string   `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
	Data struct {
		Block struct {
			Hash    string `json:"hash"`
			Number  int64  `json:"number"`
			Account struct {
				Balance string `json:"balance"`
				Nonce   string `json:"transactionCount"`
				Code    string `json:"code"`
			} `json:"account"`
		} `json:"block"`
	} `json:"data"`
}

// Balance returns the balance of a *RosettaTypes.AccountIdentifier
// at a *RosettaTypes.PartialBlockIdentifier.
//
// We must use graphql to get the balance atomically (the
// rpc method for balance does not allow for querying
// by block hash nor return the block hash where
// the balance was fetched).
func (ec *Client) Balance(
	ctx context.Context,
	account *RosettaTypes.AccountIdentifier,
	block *RosettaTypes.PartialBlockIdentifier,
) (*RosettaTypes.AccountBalanceResponse, error) {
	blockQuery := ""
	if block != nil {
		if block.Hash != nil {
			blockQuery = fmt.Sprintf(`hash: "%s"`, *block.Hash)
		}
		if block.Hash == nil && block.Index != nil {
			blockQuery = fmt.Sprintf("number: %d", *block.Index)
		}
	}

	result, err := ec.g.Query(ctx, fmt.Sprintf(`{
			block(%s){
				hash
				number
				account(address:"%s"){
					balance
					transactionCount
					code
				}
			}
		}`, blockQuery, account.Address))
	if err != nil {
		return nil, err
	}

	var bal graphqlBalance
	if err := json.Unmarshal([]byte(result), &bal); err != nil {
		return nil, err
	}

	if len(bal.Errors) > 0 {
		return nil, errors.New(RosettaTypes.PrintStruct(bal.Errors))
	}

	balance, ok := new(big.Int).SetString(bal.Data.Block.Account.Balance[2:], 16)
	if !ok {
		return nil, fmt.Errorf(
			"could not extract account balance from %s",
			bal.Data.Block.Account.Balance,
		)
	}
	nonce, ok := new(big.Int).SetString(bal.Data.Block.Account.Nonce[2:], 16)
	if !ok {
		return nil, fmt.Errorf(
			"could not extract account nonce from %s",
			bal.Data.Block.Account.Nonce,
		)
	}

	return &RosettaTypes.AccountBalanceResponse{
		Balances: []*RosettaTypes.Amount{
			{
				Value:    balance.String(),
				Currency: Currency,
			},
		},
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  bal.Data.Block.Hash,
			Index: bal.Data.Block.Number,
		},
		Metadata: map[string]interface{}{
			"nonce": nonce.Int64(),
			"code":  bal.Data.Block.Account.Code,
		},
	}, nil
}

// GetTransactionReceiptInput is the input to the call
// method "eth_getTransactionReceipt".
type GetTransactionReceiptInput struct {
	TxHash string `json:"tx_hash"`
}

// Call handles calls to the /call endpoint.
func (ec *Client) Call(
	ctx context.Context,
	request *RosettaTypes.CallRequest,
) (*RosettaTypes.CallResponse, error) {
	switch request.Method { // nolint:gocritic
	case "eth_getTransactionReceipt":
		var input GetTransactionReceiptInput
		if err := RosettaTypes.UnmarshalMap(request.Parameters, &input); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCallParametersInvalid, err.Error())
		}

		if len(input.TxHash) == 0 {
			return nil, fmt.Errorf("%w:tx_hash missing from params", ErrCallParametersInvalid)
		}

		receipt, err := ec.transactionReceipt(ctx, common.HexToHash(input.TxHash))
		if err != nil {
			return nil, err
		}

		// We cannot use RosettaTypes.MarshalMap because geth uses a custom
		// marshaler to convert *types.Receipt to JSON.
		jsonOutput, err := receipt.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCallOutputMarshal, err.Error())
		}

		var receiptMap map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &receiptMap); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCallOutputMarshal, err.Error())
		}

		// We must encode data over the wire so we can unmarshal correctly
		return &RosettaTypes.CallResponse{
			Result: receiptMap,
		}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrCallMethodInvalid, request.Method)
}
