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
	"testing"

	"github.com/coinbase/rosetta-ethereum/configuration"
	mocks "github.com/coinbase/rosetta-ethereum/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCall_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	servicer := NewCallAPIService(cfg, mockClient)
	ctx := context.Background()

	resp, err := servicer.Call(ctx, &types.CallRequest{})
	assert.Nil(t, resp)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)

	mockClient.AssertExpectations(t)
}

func TestCall_Online(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Online,
	}
	mockClient := &mocks.Client{}
	servicer := NewCallAPIService(cfg, mockClient)
	ctx := context.Background()

	request := &types.CallRequest{
		Method: "blah",
	}

	resp := &types.CallResponse{
		Result: map[string]interface{}{
			"test": "blah",
		},
	}

	mockClient.On("Call", ctx, request).Return(resp, nil).Once()
	callResp, err := servicer.Call(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, resp, callResp)

	mockClient.AssertExpectations(t)
}
