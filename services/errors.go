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
	"github.com/coinbase/rosetta-sdk-go/types"
)

var (
	// Errors contains all errors that could be returned
	// by this Rosetta implementation.
	Errors = []*types.Error{
		ErrUnimplemented,
		ErrUnavailableOffline,
		ErrGeth,
		ErrUnableToDecompressPubkey,
		ErrUnclearIntent,
		ErrUnableToParseIntermediateResult,
		ErrSignatureInvalid,
		ErrBroadcastFailed,
		ErrCallParametersInvalid,
		ErrCallOutputMarshal,
		ErrCallMethodInvalid,
		ErrBlockOrphaned,
		ErrInvalidAddress,
		ErrGethNotReady,
	}

	// ErrUnimplemented is returned when an endpoint
	// is called that is not implemented.
	ErrUnimplemented = &types.Error{
		Code:    0, //nolint
		Message: "Endpoint not implemented",
	}

	// ErrUnavailableOffline is returned when an endpoint
	// is called that is not available offline.
	ErrUnavailableOffline = &types.Error{
		Code:    1, //nolint
		Message: "Endpoint unavailable offline",
	}

	// ErrGeth is returned when geth
	// errors on a request.
	ErrGeth = &types.Error{
		Code:    2, //nolint
		Message: "geth error",
	}

	// ErrUnableToDecompressPubkey is returned when
	// the *types.PublicKey provided in /construction/derive
	// cannot be decompressed.
	ErrUnableToDecompressPubkey = &types.Error{
		Code:    3, //nolint
		Message: "unable to decompress public key",
	}

	// ErrUnclearIntent is returned when operations
	// provided in /construction/preprocess or /construction/payloads
	// are not valid.
	ErrUnclearIntent = &types.Error{
		Code:    4, //nolint
		Message: "Unable to parse intent",
	}

	// ErrUnableToParseIntermediateResult is returned
	// when a data structure passed between Construction
	// API calls is not valid.
	ErrUnableToParseIntermediateResult = &types.Error{
		Code:    5, //nolint
		Message: "Unable to parse intermediate result",
	}

	// ErrSignatureInvalid is returned when a signature
	// cannot be parsed.
	ErrSignatureInvalid = &types.Error{
		Code:    6, //nolint
		Message: "Signature invalid",
	}

	// ErrBroadcastFailed is returned when transaction
	// broadcast fails.
	ErrBroadcastFailed = &types.Error{
		Code:    7, //nolint
		Message: "Unable to broadcast transaction",
	}

	// ErrCallParametersInvalid is returned when
	// the parameters for a particular call method
	// are considered invalid.
	ErrCallParametersInvalid = &types.Error{
		Code:    8, //nolint
		Message: "Call parameters invalid",
	}

	// ErrCallOutputMarshal is returned when the output
	// for /call cannot be marshaled.
	ErrCallOutputMarshal = &types.Error{
		Code:    9, //nolint
		Message: "Call output marshal failed",
	}

	// ErrCallMethodInvalid is returned when a /call
	// method is invalid.
	ErrCallMethodInvalid = &types.Error{
		Code:    10, //nolint
		Message: "Call method invalid",
	}

	// ErrBlockOrphaned is returned when a block being
	// processed is orphaned and it is not possible
	// to gather all receipts. At some point in the future,
	// it may become possible to gather all receipts if the
	// block becomes part of the canonical chain again.
	ErrBlockOrphaned = &types.Error{
		Code:      11, //nolint
		Message:   "Block orphaned",
		Retriable: true,
	}

	// ErrInvalidAddress is returned when an address
	// is not valid.
	ErrInvalidAddress = &types.Error{
		Code:    12, //nolint
		Message: "Invalid address",
	}

	// ErrGethNotReady is returned when geth
	// cannot yet serve any queries.
	ErrGethNotReady = &types.Error{
		Code:      13, //nolint
		Message:   "geth not ready",
		Retriable: true,
	}
)

// wrapErr adds details to the types.Error provided. We use a function
// to do this so that we don't accidentially overrwrite the standard
// errors.
func wrapErr(rErr *types.Error, err error) *types.Error {
	newErr := &types.Error{
		Code:      rErr.Code,
		Message:   rErr.Message,
		Retriable: rErr.Retriable,
	}
	if err != nil {
		newErr.Details = map[string]interface{}{
			"context": err.Error(),
		}
	}

	return newErr
}
