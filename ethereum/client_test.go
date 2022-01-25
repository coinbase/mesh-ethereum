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
	"io/ioutil"
	"math/big"
	"reflect"
	"sort"
	"testing"

	mocks "github.com/coinbase/rosetta-ethereum/mocks/ethereum"

	RosettaTypes "github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/semaphore"
)

func TestStatus_NotReady(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.Nil(t, block)
	assert.Equal(t, int64(-1), timestamp)
	assert.Nil(t, syncStatus)
	assert.Nil(t, peers)
	assert.True(t, errors.Is(err, ethereum.NotFound))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_NotSyncing(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			status := args.Get(1).(*json.RawMessage)

			*status = json.RawMessage("false")
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			info := args.Get(1).(*[]*p2p.PeerInfo)

			file, err := ioutil.ReadFile("testdata/peers.json")
			assert.NoError(t, err)

			assert.NoError(t, json.Unmarshal(file, info))
		},
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0x48269a339ce1489cff6bab70eff432289c4f490b81dbd00ff1f81c68de06b842",
		Index: 8916656,
	}, block)
	assert.Equal(t, int64(1603225195000), timestamp)
	assert.Nil(t, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{
		{
			PeerID: "16dedaa93519f9ba41a50d77876aae4bfcddfa7cecf232b9abe3ab5bf0b871f3",
			Metadata: map[string]interface{}{
				"caps": []string{
					"eth/63",
					"eth/64",
					"eth/65",
				},
				"enode": "enode://5654cc39fd278c994c451434dfa7b1a44977c52018a87e911368b54daf795955d5a2dc2ece98be5a7e8d0eb245c8ef573c92e04e8b15363f9c713a8127fe7c7b@35.183.116.112:57510", // nolint
				"enr":   "",
				"name":  "Geth/v1.9.22-stable-c71a7e26/linux-amd64/go1.15",
				"protocols": map[string]interface{}{
					"eth": map[string]interface{}{
						"difficulty": float64(31779242235308530),
						"head":       "0x4a01c35e3e2627bf5a735bc9c7f336cb1e6450f93955473008ff64cf01feeef8",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "1b75a634fbc9198d73413a0ced02837707d1fd09e4e90b8b90a0abac57113299",
			Metadata: map[string]interface{}{
				"caps": []string{
					"eth/63",
					"eth/64",
					"eth/65",
				},
				"enode": "enode://bead1278155bfabdd51f04a6e896356da2f5687aa1f550bebc540828579522b87e22c67edf90efa651582e40c8c8037eb0f998208cab4a69b52c5e3387671b59@174.129.122.13:30303",                                                    // nolint
				"enr":   "enr:-Je4QICGSLfIHa7vX3bdWnKqWIS7YwmLUP6JVqU5nBhxPpH_X_Uz1pZwVS8a48uESHay1nvz9FtxLYFftpMr3wvFZJ4Qg2V0aMfGhGcn75CAgmlkgnY0gmlwhK6Beg2Jc2VjcDI1NmsxoQO-rRJ4FVv6vdUfBKboljVtovVoeqH1UL68VAgoV5UiuIN0Y3CCdl-DdWRwgnZf", // nolint
				"name":  "Geth/v1.9.15-omnibus-75eb5240/linux-amd64/go1.14.4",
				"protocols": map[string]interface{}{
					"eth": map[string]interface{}{
						"difficulty": float64(31779248439556308),
						"head":       "0x562415e43630bb6d79176ea2fa35ff2a54cee276b678b755831886b1029911bd",
						"version":    float64(65),
					},
				},
			},
		},
	}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_NotSyncing_SkipAdminCalls(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
		skipAdminCalls: true,
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			status := args.Get(1).(*json.RawMessage)

			*status = json.RawMessage("false")
		},
	).Once()

	adminPeersSkipped := true
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			adminPeersSkipped = false
		},
	).Maybe()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.True(t, adminPeersSkipped)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0x48269a339ce1489cff6bab70eff432289c4f490b81dbd00ff1f81c68de06b842",
		Index: 8916656,
	}, block)
	assert.Equal(t, int64(1603225195000), timestamp)
	assert.Nil(t, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_Syncing(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			progress := args.Get(1).(*json.RawMessage)
			file, err := ioutil.ReadFile("testdata/syncing_info.json")
			assert.NoError(t, err)

			*progress = json.RawMessage(file)
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			info := args.Get(1).(*[]*p2p.PeerInfo)

			file, err := ioutil.ReadFile("testdata/peers.json")
			assert.NoError(t, err)

			assert.NoError(t, json.Unmarshal(file, info))
		},
	).Once()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0x48269a339ce1489cff6bab70eff432289c4f490b81dbd00ff1f81c68de06b842",
		Index: 8916656,
	}, block)
	assert.Equal(t, int64(1603225195000), timestamp)
	assert.Equal(t, &RosettaTypes.SyncStatus{
		CurrentIndex: RosettaTypes.Int64(25),
		TargetIndex:  RosettaTypes.Int64(8916760),
	}, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{
		{
			PeerID: "16dedaa93519f9ba41a50d77876aae4bfcddfa7cecf232b9abe3ab5bf0b871f3",
			Metadata: map[string]interface{}{
				"caps": []string{
					"eth/63",
					"eth/64",
					"eth/65",
				},
				"enode": "enode://5654cc39fd278c994c451434dfa7b1a44977c52018a87e911368b54daf795955d5a2dc2ece98be5a7e8d0eb245c8ef573c92e04e8b15363f9c713a8127fe7c7b@35.183.116.112:57510", // nolint
				"enr":   "",
				"name":  "Geth/v1.9.22-stable-c71a7e26/linux-amd64/go1.15",
				"protocols": map[string]interface{}{
					"eth": map[string]interface{}{
						"difficulty": float64(31779242235308530),
						"head":       "0x4a01c35e3e2627bf5a735bc9c7f336cb1e6450f93955473008ff64cf01feeef8",
						"version":    float64(65),
					},
				},
			},
		},
		{
			PeerID: "1b75a634fbc9198d73413a0ced02837707d1fd09e4e90b8b90a0abac57113299",
			Metadata: map[string]interface{}{
				"caps": []string{
					"eth/63",
					"eth/64",
					"eth/65",
				},
				"enode": "enode://bead1278155bfabdd51f04a6e896356da2f5687aa1f550bebc540828579522b87e22c67edf90efa651582e40c8c8037eb0f998208cab4a69b52c5e3387671b59@174.129.122.13:30303",                                                    // nolint
				"enr":   "enr:-Je4QICGSLfIHa7vX3bdWnKqWIS7YwmLUP6JVqU5nBhxPpH_X_Uz1pZwVS8a48uESHay1nvz9FtxLYFftpMr3wvFZJ4Qg2V0aMfGhGcn75CAgmlkgnY0gmlwhK6Beg2Jc2VjcDI1NmsxoQO-rRJ4FVv6vdUfBKboljVtovVoeqH1UL68VAgoV5UiuIN0Y3CCdl-DdWRwgnZf", // nolint
				"name":  "Geth/v1.9.15-omnibus-75eb5240/linux-amd64/go1.14.4",
				"protocols": map[string]interface{}{
					"eth": map[string]interface{}{
						"difficulty": float64(31779248439556308),
						"head":       "0x562415e43630bb6d79176ea2fa35ff2a54cee276b678b755831886b1029911bd",
						"version":    float64(65),
					},
				},
			},
		},
	}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestStatus_Syncing_SkipAdminCalls(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
		skipAdminCalls: true,
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			header := args.Get(1).(**types.Header)
			file, err := ioutil.ReadFile("testdata/basic_header.json")
			assert.NoError(t, err)

			*header = new(types.Header)

			assert.NoError(t, (*header).UnmarshalJSON(file))
		},
	).Once()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_syncing",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			progress := args.Get(1).(*json.RawMessage)
			file, err := ioutil.ReadFile("testdata/syncing_info.json")
			assert.NoError(t, err)

			*progress = json.RawMessage(file)
		},
	).Once()

	adminPeersSkipped := true
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"admin_peers",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			adminPeersSkipped = false
		},
	).Maybe()

	block, timestamp, syncStatus, peers, err := c.Status(ctx)
	assert.True(t, adminPeersSkipped)
	assert.Equal(t, &RosettaTypes.BlockIdentifier{
		Hash:  "0x48269a339ce1489cff6bab70eff432289c4f490b81dbd00ff1f81c68de06b842",
		Index: 8916656,
	}, block)
	assert.Equal(t, int64(1603225195000), timestamp)
	assert.Equal(t, &RosettaTypes.SyncStatus{
		CurrentIndex: RosettaTypes.Int64(25),
		TargetIndex:  RosettaTypes.Int64(8916760),
	}, syncStatus)
	assert.Equal(t, []*RosettaTypes.Peer{}, peers)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x4cfc400fed52f9681b42454c2db4b18ab98f8de1.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		nil,
	)
	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0x9999286598edf07606228ba0233736e544a086a8822c61f9db3706887fc25dda",
			Index: 8165,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "10372550232136640000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_Historical_Hash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x4cfc400fed52f9681b42454c2db4b18ab98f8de1.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(hash: "0x9999286598edf07606228ba0233736e544a086a8822c61f9db3706887fc25dda"){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0x9999286598edf07606228ba0233736e544a086a8822c61f9db3706887fc25dda",
			),
			Index: RosettaTypes.Int64(8165),
		},
	)
	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0x9999286598edf07606228ba0233736e544a086a8822c61f9db3706887fc25dda",
			Index: 8165,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "10372550232136640000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_Historical_Index(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile(
		"testdata/account_balance_0x4cfc400fed52f9681b42454c2db4b18ab98f8de1.json",
	)
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(number: 8165){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(8165),
		},
	)
	assert.Equal(t, &RosettaTypes.AccountBalanceResponse{
		BlockIdentifier: &RosettaTypes.BlockIdentifier{
			Hash:  "0x9999286598edf07606228ba0233736e544a086a8822c61f9db3706887fc25dda",
			Index: 8165,
		},
		Balances: []*RosettaTypes.Amount{
			{
				Value:    "10372550232136640000000",
				Currency: Currency,
			},
		},
		Metadata: map[string]interface{}{
			"code":  "0x",
			"nonce": int64(0),
		},
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_InvalidAddress(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile("testdata/account_balance_invalid.json")
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(){
				hash
				number
				account(address:"0x4cfc400fed52f9681b42454c2db4b18ab98f8de"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x4cfc400fed52f9681b42454c2db4b18ab98f8de",
		},
		nil,
	)
	assert.Nil(t, resp)
	assert.Error(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBalance_InvalidHash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	result, err := ioutil.ReadFile("testdata/account_balance_invalid_block.json")
	assert.NoError(t, err)
	mockGraphQL.On(
		"Query",
		ctx,
		`{
			block(hash: "0x7d2a2713026a0e66f131878de2bb2df2fff6c24562c1df61ec0265e5fedf2626"){
				hash
				number
				account(address:"0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55"){
					balance
					transactionCount
					code
				}
			}
		}`,
	).Return(
		string(result),
		nil,
	).Once()

	resp, err := c.Balance(
		ctx,
		&RosettaTypes.AccountIdentifier{
			Address: "0x2f93B2f047E05cdf602820Ac4B3178efc2b43D55",
		},
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0x7d2a2713026a0e66f131878de2bb2df2fff6c24562c1df61ec0265e5fedf2626",
			),
		},
	)
	assert.Nil(t, resp)
	assert.Error(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetBlockByNumber(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2af0",
		false,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*map[string]interface{})

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, r)
			assert.NoError(t, err)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_10992.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getBlockByNumber",
			Parameters: map[string]interface{}{
				"index":                    RosettaTypes.Int64(10992),
				"show_transaction_details": false,
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetBlockByNumber_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getBlockByNumber",
			Parameters: map[string]interface{}{
				"index":                    "a string",
				"show_transaction_details": false,
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetTransactionReceipt(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionReceipt",
		common.HexToHash("0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d"),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(**types.Receipt)

			file, err := ioutil.ReadFile(
				"testdata/call_0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d.json",
			)
			assert.NoError(t, err)

			*r = new(types.Receipt)

			assert.NoError(t, (*r).UnmarshalJSON(file))
		},
	).Once()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getTransactionReceipt",
			Parameters: map[string]interface{}{
				"tx_hash": "0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d",
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result: map[string]interface{}{
			"blockHash":         "0x928b4d7d1ab8fcb2f62ffa7bba7a1a52251a1145ffc0faec3e009535ba4a2669",
			"blockNumber":       "0x7edcff",
			"contractAddress":   "0x0000000000000000000000000000000000000000",
			"cumulativeGasUsed": "0x744f1b",
			"gasUsed":           "0x5208",
			"logs":              []interface{}{},
			"logsBloom":         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", // nolint
			"root":              "0x",
			"status":            "0x1",
			"transactionHash":   "0xb358c6958b1cab722752939cbb92e3fec6b6023de360305910ce80c56c3dad9d",
			"transactionIndex":  "0x21",
		},
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_GetTransactionReceipt_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_getTransactionReceipt",
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_Call(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_call",
		map[string]string{
			"to":   "0xB5E5D0F8C0cbA267CD3D7035d6AdC8eBA7Df7Cdd",
			"data": "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
		},
		toBlockNumArg(big.NewInt(11408349)),
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*string)

			var expected map[string]interface{}
			file, err := ioutil.ReadFile("testdata/call_balance_11408349.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, &expected)
			assert.NoError(t, err)

			*r = expected["data"].(string)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/call_balance_11408349.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_call",
			Parameters: map[string]interface{}{
				"index": 11408349,
				"to":    "0xB5E5D0F8C0cbA267CD3D7035d6AdC8eBA7Df7Cdd",
				"data":  "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_Call_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_call",
			Parameters: map[string]interface{}{
				"index": 11408349,
				"Hash":  "0x73fc065bc04f16c98247f8ec1e990f581ec58723bcd8059de85f93ab18706448",
				"to":    "not valid  ",
				"data":  "0x70a08231000000000000000000000000b5e5d0f8c0cba267cd3d7035d6adc8eba7df7cdd",
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_EstimateGas(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()

	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_estimateGas",
		map[string]string{
			"from": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
			"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
			"data": "0xa9059cbb000000000000000000000000ae7e48ee0f758cd706b76cf7e2175d982800879a" +
				"00000000000000000000000000000000000000000000000000521c5f98b8ea00",
		},
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*string)

			var expected map[string]interface{}
			file, err := ioutil.ReadFile("testdata/estimate_gas_0xaD6D458402F60fD3Bd25163575031ACDce07538D.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, &expected)
			assert.NoError(t, err)

			*r = expected["data"].(string)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/estimate_gas_0xaD6D458402F60fD3Bd25163575031ACDce07538D.json")
	assert.NoError(t, err)
	var correct map[string]interface{}
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_estimateGas",
			Parameters: map[string]interface{}{
				"from": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
				"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
				"data": "0xa9059cbb000000000000000000000000ae7e48ee0f758cd706b76cf7e2175d982800879a" +
					"00000000000000000000000000000000000000000000000000521c5f98b8ea00",
			},
		},
	)
	assert.Equal(t, &RosettaTypes.CallResponse{
		Result:     correct,
		Idempotent: false,
	}, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_EstimateGas_InvalidArgs(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "eth_estimateGas",
			Parameters: map[string]interface{}{
				"From": "0xE550f300E477C60CE7e7172d12e5a27e9379D2e3",
				"to":   "0xaD6D458402F60fD3Bd25163575031ACDce07538D",
			},
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallParametersInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestCall_InvalidMethod(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	resp, err := c.Call(
		ctx,
		&RosettaTypes.CallRequest{
			Method: "blah",
		},
	)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, ErrCallMethodInvalid))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func testTraceConfig() (*tracers.TraceConfig, error) {
	loadedFile, err := ioutil.ReadFile("call_tracer.js")
	if err != nil {
		return nil, fmt.Errorf("%w: could not load tracer file", err)
	}

	loadedTracer := string(loadedFile)
	return &tracers.TraceConfig{
		Timeout: &tracerTimeout,
		Tracer:  &loadedTracer,
	}, nil
}

func TestBlock_Current(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"latest",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		nil,
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_Hash(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByHash",
		"0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Hash: RosettaTypes.String(
				"0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae",
			),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_Index(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2af0",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10992.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xba9ded5ca1ec9adb9451bf062c9de309d9552fa0f0254a7b982d3daf7ae436ae.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10992.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(10992),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestBlock_FirstBlock(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x0",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_0.json")
			assert.NoError(t, err)

			*r = file
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_0.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(0),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func jsonifyBlock(b *RosettaTypes.Block) (*RosettaTypes.Block, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	var bo RosettaTypes.Block
	if err := json.Unmarshal(bytes, &bo); err != nil {
		return nil, err
	}

	return &bo, nil
}

// Block with transaction
func TestBlock_10994(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2af2",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10994.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xb6a2558c2e54bfb11247d0764311143af48d122f29fc408d9519f47d70aa2d50"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xb6a2558c2e54bfb11247d0764311143af48d122f29fc408d9519f47d70aa2d50.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 1)
			assert.Equal(
				t,
				"0xd83b1dcf7d47c4115d78ce0361587604e8157591b118bd64ada02e86c9d5ca7e",
				r[0].Args[0],
			)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0xd83b1dcf7d47c4115d78ce0361587604e8157591b118bd64ada02e86c9d5ca7e.json",
			) // nolint
			assert.NoError(t, err)

			receipt := new(types.Receipt)
			assert.NoError(t, receipt.UnmarshalJSON(file))
			*(r[0].Result.(**types.Receipt)) = receipt
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10994.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(10994),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with uncle
func TestBlock_10991(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x2aef",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_10991.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0x4cd21f49705529e2628f8ae1a248bcd0e3cafd21bf6d741bdee2820af82cff95"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0x4cd21f49705529e2628f8ae1a248bcd0e3cafd21bf6d741bdee2820af82cff95.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 1)
			assert.Equal(
				t,
				common.HexToHash(
					"0x4cd21f49705529e2628f8ae1a248bcd0e3cafd21bf6d741bdee2820af82cff95",
				),
				r[0].Args[0],
			)
			assert.Equal(t, "0x0", r[0].Args[1])

			file, err := ioutil.ReadFile(
				"testdata/uncle_0x8e585e32e6beb4b1f60377d53210a521ace5c30395c34398d535ea56edcf8899.json",
			) // nolint
			assert.NoError(t, err)

			header := new(types.Header)
			assert.NoError(t, header.UnmarshalJSON(file))
			*(r[0].Result.(**types.Header)) = header
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_10991.json")
	assert.NoError(t, err)
	var correct *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correct))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(10991),
		},
	)
	assert.Equal(t, correct.Block, resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with partial success transaction
func TestBlock_239782(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x3a8a6",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_239782.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xc4487850a40d85b79cf5e5b69db38284fbd39efcf902ca8a6d9f2ba89c538ea3"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xc4487850a40d85b79cf5e5b69db38284fbd39efcf902ca8a6d9f2ba89c538ea3.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 1)
			assert.Equal(
				t,
				"0x05613760334d347e771fad61b1815c8c817b8dd5f0fcbba57c3f2df67dec33d6",
				r[0].Args[0],
			)

			file, err := ioutil.ReadFile(
				"testdata/tx_receipt_0x05613760334d347e771fad61b1815c8c817b8dd5f0fcbba57c3f2df67dec33d6.json",
			) // nolint
			assert.NoError(t, err)

			receipt := new(types.Receipt)
			assert.NoError(t, receipt.UnmarshalJSON(file))
			*(r[0].Result.(**types.Receipt)) = receipt
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_239782.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(239782),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with transfer to destroyed contract
func TestBlock_363415(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x58b97",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_363415.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xf0445269b02ba461af662d8c6aac50d9557a0cc9dbe580d3e180efd7879cc79e"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xf0445269b02ba461af662d8c6aac50d9557a0cc9dbe580d3e180efd7879cc79e.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 2)

			for i, txHash := range []string{
				"0x9e0f7c64a5bf1fc9f3d7b7963cf23f74e3d2c0b2b3f35f26df031954e5581179",
				"0x0046a7c3ca126864a3e851235ca6bf030300f9138f035f5f190e59ff9a4b22ff",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_363415.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(363415),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with transfer to precompiled
func TestBlock_363753(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x58ce9",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_363753.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0x3defb56cc49cf7603e08749516a003baae0944596e4555b0d868ec225ff2bcd3"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0x3defb56cc49cf7603e08749516a003baae0944596e4555b0d868ec225ff2bcd3.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 2)

			for i, txHash := range []string{
				"0x586d0a158f29da3d0e8fa4d24596d1a9f6ded03b5ccdb68f40e9372980488fc8",
				"0x80fb7e6bfa8dae67cf79f21b9e68c5af727ba52f3ab1e5a5be5c8048a9758f56",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_363753.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(363753),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with complex self-destruct
func TestBlock_468179(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x724d3",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_468179.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xd88e8376ec3eef899d9fbc6349e8330ebfc102b245fef784a999ac854091cb64"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xd88e8376ec3eef899d9fbc6349e8330ebfc102b245fef784a999ac854091cb64.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 2)

			for i, txHash := range []string{
				"0x712f7aed1ac12f8a38b4caefea8e7c1940c88add78e110b194c653c9efb3a75d",
				"0x99b723ac54002b16049143474d80f8e6358d14dec2250d873511d091de74977d",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_468179.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(468179),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with complex resurrection
func TestBlock_363366(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x58b66",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_363366.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0x5f7c67c2eb0e828b0f4a0e64d5fbae0ed66b70c9ae752e6175c9ef62402502df"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0x5f7c67c2eb0e828b0f4a0e64d5fbae0ed66b70c9ae752e6175c9ef62402502df.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 3)

			for i, txHash := range []string{
				"0x3f11ca203c7fd814751725c2c5a3efa00bebbbd5e89f406a28b4a36559393b6f",
				"0x4cc86d845b6ee5c12db00cc75c42e98f8bbf62060bc925942c5ff6a36878549b",
				"0xf8b84ff00db596c9db15de1a44c939cce36c0dfd60ef6171db6951b11d7d015d",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_363366.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(363366),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

// Block with blackholed funds
func TestBlock_468194(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	tc, err := testTraceConfig()
	assert.NoError(t, err)
	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		tc:             tc,
		p:              params.RopstenChainConfig,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getBlockByNumber",
		"0x724e2",
		true,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile("testdata/block_468194.json")
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"debug_traceBlockByHash",
		common.HexToHash("0xf0d9ab47473e38f98b195ba7a17934f68519168f5fdec9899b3c18180d8fbb54"),
		tc,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*json.RawMessage)

			file, err := ioutil.ReadFile(
				"testdata/block_trace_0xf0d9ab47473e38f98b195ba7a17934f68519168f5fdec9899b3c18180d8fbb54.json",
			) // nolint
			assert.NoError(t, err)

			*r = json.RawMessage(file)
		},
	).Once()
	mockJSONRPC.On(
		"BatchCallContext",
		ctx,
		mock.Anything,
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).([]rpc.BatchElem)

			assert.Len(t, r, 2)

			for i, txHash := range []string{
				"0xbd54f0c5742a5c96ffb358680b88a0f6cfbf83d599dbd0b8fff66b59ed0d7f81",
				"0xf3626ec6a7aba22137b012e8e68513dcaf8574d0412b97e4381513a3ca9ecfc0",
			} {
				assert.Equal(
					t,
					txHash,
					r[i].Args[0],
				)

				file, err := ioutil.ReadFile(
					"testdata/tx_receipt_" + txHash + ".json",
				) // nolint
				assert.NoError(t, err)

				receipt := new(types.Receipt)
				assert.NoError(t, receipt.UnmarshalJSON(file))
				*(r[i].Result.(**types.Receipt)) = receipt
			}
		},
	).Once()

	correctRaw, err := ioutil.ReadFile("testdata/block_response_468194.json")
	assert.NoError(t, err)
	var correctResp *RosettaTypes.BlockResponse
	assert.NoError(t, json.Unmarshal(correctRaw, &correctResp))

	resp, err := c.Block(
		ctx,
		&RosettaTypes.PartialBlockIdentifier{
			Index: RosettaTypes.Int64(468194),
		},
	)
	assert.NoError(t, err)

	// Ensure types match
	jsonResp, err := jsonifyBlock(resp)
	assert.NoError(t, err)
	assert.Equal(t, correctResp.Block, jsonResp)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestPendingNonceAt(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_getTransactionCount",
		common.HexToAddress("0xfFC614eE978630D7fB0C06758DeB580c152154d3"),
		"pending",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*hexutil.Uint64)

			*r = hexutil.Uint64(10)
		},
	).Once()
	resp, err := c.PendingNonceAt(
		ctx,
		common.HexToAddress("0xfFC614eE978630D7fB0C06758DeB580c152154d3"),
	)
	assert.Equal(t, uint64(10), resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestSuggestGasPrice(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_gasPrice",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r := args.Get(1).(*hexutil.Big)

			*r = *(*hexutil.Big)(big.NewInt(100000))
		},
	).Once()
	resp, err := c.SuggestGasPrice(
		ctx,
	)
	assert.Equal(t, big.NewInt(100000), resp)
	assert.NoError(t, err)

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestSendTransaction(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	ctx := context.Background()
	mockJSONRPC.On(
		"CallContext",
		ctx,
		mock.Anything,
		"eth_sendRawTransaction",
		"0xf86a80843b9aca00825208941ff502f9fe838cd772874cb67d0d96b93fd1d6d78725d4b6199a415d8029a01d110bf9fd468f7d00b3ce530832e99818835f45e9b08c66f8d9722264bb36c7a02711f47ec99f9ac585840daef41b7118b52ec72f02fcb30d874d36b10b668b59", // nolint
	).Return(
		nil,
	).Once()

	rawTx, err := ioutil.ReadFile("testdata/submitted_tx.json")
	assert.NoError(t, err)

	tx := new(types.Transaction)
	assert.NoError(t, tx.UnmarshalJSON(rawTx))

	assert.NoError(t, c.SendTransaction(
		ctx,
		tx,
	))

	mockJSONRPC.AssertExpectations(t)
	mockGraphQL.AssertExpectations(t)
}

func TestGetMempool(t *testing.T) {
	mockJSONRPC := &mocks.JSONRPC{}
	mockGraphQL := &mocks.GraphQL{}
	ctx := context.Background()
	expectedMempool := &RosettaTypes.MempoolResponse{
		TransactionIdentifiers: []*RosettaTypes.TransactionIdentifier{
			{Hash: "0x994024ef9f05d1cb25d01572642c1f550c78d214a52c306bb100d22c025b59d4"},
			{Hash: "0x0859cb844087a280fa031ce2a0879f4f9832f431d6de67f57d0ca32e90dd9e21"},
			{Hash: "0xa6de13e0a4465c9b55726d0826d020ed179fa1bda882a00e90aa467266af1815"},
			{Hash: "0x94e28f01121fd56dead7b89c46c8d6ba32bb79dad5f7e29c1d523224479f10f4"},
			{Hash: "0xf98b8a6589fa27f66545f31d443140c16eaa3da3896e66c4f05a03defa12cda4"},
			{Hash: "0x4c652b9e779277f06e17b6a4c4db6658ac012f7c93f5eeab80e7802cce3ff556"},
			{Hash: "0xa81f636c4b47831efd324dce5c48fe3a3d7a4f0020fae063887e8c2765ff1088"},
			{Hash: "0x3bc19098a49974002ba933ac8f9c753ce5af538ab93fc68586849e8b8b0dce80"},
			{Hash: "0x3401ec802cbe4b02d5be18717b53bb8177bd746f95a54da4674d7c27620facda"},
			{Hash: "0x3d1c58eced8f15f3224e9b2579420d078dbd2adba8e825cee91fd306b0943f89"},
			{Hash: "0xda591f0b15423aedb52f6b0e778b1fbc6757547d69277e2ba1aa7093583d1efb"},
			{Hash: "0xb39652e66c57a6693e44b92b5c471952c26cabf9442a9a76c372f8690658cb4c"},
			{Hash: "0x1e8700bf7215b2da0cfadcf34717f387577254e0871d828e5453a0593ce8060f"},
			{Hash: "0x83811383fb7840a03c25a7cbae7e9af138b17853563eb9e212727be2d0b9667f"},
			{Hash: "0x1e53751e1312cae3324a6b36c67dc95bfec993d7b4939c0de8c0dc761a0afd31"},
			{Hash: "0xc1052f9378db5a779c42ae2de9a0b94c8a6357c815446d6ba55485dcc1b187ef"},
		},
	}

	c := &Client{
		c:              mockJSONRPC,
		g:              mockGraphQL,
		traceSemaphore: semaphore.NewWeighted(100),
	}

	mockJSONRPC.On(
		"CallContext", ctx, mock.Anything, "txpool_content",
	).Return(
		nil,
	).Run(
		func(args mock.Arguments) {
			r, ok := args.Get(1).(*TxPoolContentResponse)
			assert.True(t, ok)

			file, err := ioutil.ReadFile("testdata/txpool_content.json")
			assert.NoError(t, err)

			err = json.Unmarshal(file, r)
			assert.NoError(t, err)
		},
	).Once()

	actualMempool, err := c.GetMempool(ctx)
	assert.NoError(t, err)

	assert.Len(t, actualMempool.TransactionIdentifiers, len(expectedMempool.TransactionIdentifiers))

	// Sort both slices to compare later
	sort.Slice(expectedMempool.TransactionIdentifiers, func(i, j int) bool {
		return expectedMempool.TransactionIdentifiers[i].Hash < expectedMempool.TransactionIdentifiers[j].Hash
	})

	sort.Slice(actualMempool.TransactionIdentifiers, func(i, j int) bool {
		return actualMempool.TransactionIdentifiers[i].Hash < actualMempool.TransactionIdentifiers[j].Hash
	})

	assert.True(t, reflect.DeepEqual(actualMempool, expectedMempool))

	mockJSONRPC.AssertExpectations(t)
}
