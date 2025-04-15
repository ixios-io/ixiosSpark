// Copyright 2021 The go-ethereum Authors
// This file is part of the IxiosSpark library, which builds upon the source code of the go-ethereum library.
//
// The IxiosSpark library, including the go-ethereum library source code it is based on, is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The IxiosSpark and go-ethereum library source code are distributed with the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
	"github.com/ixios-io/ixiosSpark/common/math"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/crypto/kzg4844"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/ixios-io/ixiosSpark/rpc"
)

var (
	maxBlobsPerTransaction = params.MaxBlobGasPerBlock / params.BlobTxBlobGasPerBlob
)

// TransactionArgs represents the arguments to construct a new transaction
// or a message call.
type TransactionArgs struct {
	From                 *common.Address `json:"from"`
	To                   *common.Address `json:"to"`
	Gas                  *hexutil.Uint64 `json:"gas"`
	GasPrice             *hexutil.Big    `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas"`
	Value                *hexutil.Big    `json:"value"`
	Nonce                *hexutil.Uint64 `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://archive.is/gdG8p
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`

	// Introduced by AccessListTxType transaction.
	AccessList *types.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big      `json:"chainId,omitempty"`

	// For BlobTxType
	BlobFeeCap *hexutil.Big  `json:"maxFeePerBlobGas"`
	BlobHashes []common.Hash `json:"blobVersionedHashes,omitempty"`

	// For BlobTxType transactions with blob sidecar
	Blobs       []kzg4844.Blob       `json:"blobs"`
	Commitments []kzg4844.Commitment `json:"commitments"`
	Proofs      []kzg4844.Proof      `json:"proofs"`

	// This configures whether blobs are allowed to be passed.
	blobSidecarAllowed bool
}

// from retrieves the transaction sender address.
func (args *TransactionArgs) from() common.Address {
	if args.From == nil {
		return common.Address{}
	}
	return *args.From
}

// data retrieves the transaction calldata. Input field is preferred.
func (args *TransactionArgs) data() []byte {
	if args.Input != nil {
		return *args.Input
	}
	if args.Data != nil {
		return *args.Data
	}
	return nil
}

// setDefaults fills in default values for unspecified tx fields.
func (args *TransactionArgs) setDefaults(ctx context.Context, b Backend, skipGasEstimation bool) error {
	if err := args.setBlobTxSidecar(ctx, b); err != nil {
		return err
	}
	if err := args.setFeeDefaults(ctx, b); err != nil {
		return err
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce, err := b.GetPoolNonce(ctx, args.from())
		if err != nil {
			return err
		}
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`both "data" and "input" are set and not equal. Please use "input" to pass transaction call data`)
	}

	// BlobTx fields
	if args.BlobHashes != nil && len(args.BlobHashes) == 0 {
		return errors.New(`need at least 1 blob for a blob transaction`)
	}
	if args.BlobHashes != nil && len(args.BlobHashes) > maxBlobsPerTransaction {
		return fmt.Errorf(`too many blobs in transaction (have=%d, max=%d)`, len(args.BlobHashes), maxBlobsPerTransaction)
	}

	// create check
	if args.To == nil {
		if args.BlobHashes != nil {
			return errors.New(`missing "to" in blob transaction`)
		}
		if len(args.data()) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}

	if args.Gas == nil {
		if skipGasEstimation { // Skip gas usage estimation if a precise gas limit is not critical, e.g., in non-transaction calls.
			gas := hexutil.Uint64(b.RPCGasCap())
			if gas == 0 {
				gas = hexutil.Uint64(math.MaxUint64 / 2)
			}
			args.Gas = &gas
		} else { // Estimate the gas usage otherwise.
			// These fields are immutable during the estimation, safe to
			// pass the pointer directly.
			data := args.data()
			callArgs := TransactionArgs{
				From:                 args.From,
				To:                   args.To,
				GasPrice:             args.GasPrice,
				MaxFeePerGas:         args.MaxFeePerGas,
				MaxPriorityFeePerGas: args.MaxPriorityFeePerGas,
				Value:                args.Value,
				Data:                 (*hexutil.Bytes)(&data),
				AccessList:           args.AccessList,
				BlobFeeCap:           args.BlobFeeCap,
				BlobHashes:           args.BlobHashes,
			}
			latestBlockNr := rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)
			estimated, err := DoEstimateGas(ctx, b, callArgs, latestBlockNr, nil, b.RPCGasCap())
			if err != nil {
				return err
			}
			args.Gas = &estimated
			log.Trace("Estimate gas usage automatically", "gas", args.Gas)
		}
	}

	// If chain id is provided, ensure it matches the local chain id. Otherwise, set the local
	// chain id as the default.
	want := b.ChainConfig().ChainID
	if args.ChainID != nil {
		if have := (*big.Int)(args.ChainID); have.Cmp(want) != 0 {
			return fmt.Errorf("chainId does not match node's (have=%v, want=%v)", have, want)
		}
	} else {
		args.ChainID = (*hexutil.Big)(want)
	}
	return nil
}

// setFeeDefaults fills in default fee values for unspecified tx fields.
func (args *TransactionArgs) setFeeDefaults(ctx context.Context, b Backend) error {
	// We're always in pre-London mode, so just set gas price
	if args.GasPrice == nil {
		price, err := b.SuggestGasTipCap(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	}

	// Make sure these are nil to avoid London-style processing
	args.MaxFeePerGas = nil
	args.MaxPriorityFeePerGas = nil
	return nil
}

// setLondonFeeDefaults fills in reasonable default fee values for unspecified fields.
func (args *TransactionArgs) setLondonFeeDefaults(ctx context.Context, head *types.Header, b Backend) error {
	// Set maxPriorityFeePerGas if it is missing.
	if args.MaxPriorityFeePerGas == nil {
		tip, err := b.SuggestGasTipCap(ctx)
		if err != nil {
			return err
		}
		args.MaxPriorityFeePerGas = (*hexutil.Big)(tip)
	}
	// Set maxFeePerGas if it is missing.
	if args.MaxFeePerGas == nil {
		// Set the max fee to be 2 times larger than the previous block's base fee.
		// The additional slack allows the tx to not become invalidated if the base
		// fee is rising.
		val := new(big.Int).Add(
			args.MaxPriorityFeePerGas.ToInt(),
			new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
		)
		args.MaxFeePerGas = (*hexutil.Big)(val)
	}
	// Both EIP-1559 fee parameters are now set; sanity check them.
	if args.MaxFeePerGas.ToInt().Cmp(args.MaxPriorityFeePerGas.ToInt()) < 0 {
		return fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", args.MaxFeePerGas, args.MaxPriorityFeePerGas)
	}
	return nil
}

// setBlobTxSidecar adds the blob tx
func (args *TransactionArgs) setBlobTxSidecar(ctx context.Context, b Backend) error {
	// No blobs, we're done.
	if args.Blobs == nil {
		return nil
	}

	// Passing blobs is not allowed in all contexts, only in specific methods.
	if !args.blobSidecarAllowed {
		return errors.New(`"blobs" is not supported for this RPC method`)
	}

	n := len(args.Blobs)
	// Assume user provides either only blobs (w/o hashes), or
	// blobs together with commitments and proofs.
	if args.Commitments == nil && args.Proofs != nil {
		return errors.New(`blob proofs provided while commitments were not`)
	} else if args.Commitments != nil && args.Proofs == nil {
		return errors.New(`blob commitments provided while proofs were not`)
	}

	// len(blobs) == len(commitments) == len(proofs) == len(hashes)
	if args.Commitments != nil && len(args.Commitments) != n {
		return fmt.Errorf("number of blobs and commitments mismatch (have=%d, want=%d)", len(args.Commitments), n)
	}
	if args.Proofs != nil && len(args.Proofs) != n {
		return fmt.Errorf("number of blobs and proofs mismatch (have=%d, want=%d)", len(args.Proofs), n)
	}
	if args.BlobHashes != nil && len(args.BlobHashes) != n {
		return fmt.Errorf("number of blobs and hashes mismatch (have=%d, want=%d)", len(args.BlobHashes), n)
	}

	if args.Commitments == nil {
		// Generate commitment and proof.
		commitments := make([]kzg4844.Commitment, n)
		proofs := make([]kzg4844.Proof, n)
		for i, b := range args.Blobs {
			c, err := kzg4844.BlobToCommitment(b)
			if err != nil {
				return fmt.Errorf("blobs[%d]: error computing commitment: %v", i, err)
			}
			commitments[i] = c
			p, err := kzg4844.ComputeBlobProof(b, c)
			if err != nil {
				return fmt.Errorf("blobs[%d]: error computing proof: %v", i, err)
			}
			proofs[i] = p
		}
		args.Commitments = commitments
		args.Proofs = proofs
	} else {
		for i, b := range args.Blobs {
			if err := kzg4844.VerifyBlobProof(b, args.Commitments[i], args.Proofs[i]); err != nil {
				return fmt.Errorf("failed to verify blob proof: %v", err)
			}
		}
	}

	hashes := make([]common.Hash, n)
	hasher := sha256.New()
	for i, c := range args.Commitments {
		hashes[i] = kzg4844.CalcBlobHashV1(hasher, &c)
	}
	if args.BlobHashes != nil {
		for i, h := range hashes {
			if h != args.BlobHashes[i] {
				return fmt.Errorf("blob hash verification failed (have=%s, want=%s)", args.BlobHashes[i], h)
			}
		}
	} else {
		args.BlobHashes = hashes
	}
	return nil
}

// ToMessage converts the transaction arguments to the Message type used by the
// core evm. This method is used in calls and traces that do not require a real
// live transaction.
func (args *TransactionArgs) ToMessage(globalGasCap uint64, baseFee *big.Int) (*core.Message, error) {
	// Reject invalid combinations of pre- and post-1559 fee styles
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return nil, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}
	// Set sender address or use zero address if none specified.
	addr := args.from()

	// Set default gas & gas price if none were set
	gas := globalGasCap
	if gas == 0 {
		gas = uint64(math.MaxUint64 / 2)
	}
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != 0 && globalGasCap < gas {
		log.Warn("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
		gas = globalGasCap
	}
	var (
		gasPrice   *big.Int
		gasFeeCap  *big.Int
		gasTipCap  *big.Int
		blobFeeCap *big.Int
	)
	if baseFee == nil {
		// If there's no basefee, then it must be a non-1559 execution
		gasPrice = new(big.Int)
		if args.GasPrice != nil {
			gasPrice = args.GasPrice.ToInt()
		}
		gasFeeCap, gasTipCap = gasPrice, gasPrice
	} else {
		// A basefee is provided, necessitating 1559-type execution
		if args.GasPrice != nil {
			// User specified the legacy gas field, convert to 1559 gas typing
			gasPrice = args.GasPrice.ToInt()
			gasFeeCap, gasTipCap = gasPrice, gasPrice
		} else {
			// User specified 1559 gas fields (or none), use those
			gasFeeCap = new(big.Int)
			if args.MaxFeePerGas != nil {
				gasFeeCap = args.MaxFeePerGas.ToInt()
			}
			gasTipCap = new(big.Int)
			if args.MaxPriorityFeePerGas != nil {
				gasTipCap = args.MaxPriorityFeePerGas.ToInt()
			}
			// Backfill the legacy gasPrice for EVM execution, unless we're all zeroes
			gasPrice = new(big.Int)
			if gasFeeCap.BitLen() > 0 || gasTipCap.BitLen() > 0 {
				gasPrice = math.BigMin(new(big.Int).Add(gasTipCap, baseFee), gasFeeCap)
			}
		}
	}
	if args.BlobFeeCap != nil {
		blobFeeCap = args.BlobFeeCap.ToInt()
	} else if args.BlobHashes != nil {
		blobFeeCap = new(big.Int)
	}
	value := new(big.Int)
	if args.Value != nil {
		value = args.Value.ToInt()
	}
	data := args.data()
	var accessList types.AccessList
	if args.AccessList != nil {
		accessList = *args.AccessList
	}
	msg := &core.Message{
		From:              addr,
		To:                args.To,
		Value:             value,
		GasLimit:          gas,
		GasPrice:          gasPrice,
		GasFeeCap:         gasFeeCap,
		GasTipCap:         gasTipCap,
		Data:              data,
		AccessList:        accessList,
		BlobGasFeeCap:     blobFeeCap,
		BlobHashes:        args.BlobHashes,
		SkipAccountChecks: true,
	}
	return msg, nil
}

// toTransaction converts the arguments to a transaction.
// This assumes that setDefaults has been called.
func (args *TransactionArgs) toTransaction() *types.Transaction {
	var data types.TxData
	switch {
	case args.BlobHashes != nil:
		al := types.AccessList{}
		if args.AccessList != nil {
			al = *args.AccessList
		}
		data = &types.BlobTx{
			To:         *args.To,
			ChainID:    uint256.MustFromBig((*big.Int)(args.ChainID)),
			Nonce:      uint64(*args.Nonce),
			Gas:        uint64(*args.Gas),
			GasFeeCap:  uint256.MustFromBig((*big.Int)(args.MaxFeePerGas)),
			GasTipCap:  uint256.MustFromBig((*big.Int)(args.MaxPriorityFeePerGas)),
			Value:      uint256.MustFromBig((*big.Int)(args.Value)),
			Data:       args.data(),
			AccessList: al,
			BlobHashes: args.BlobHashes,
			BlobFeeCap: uint256.MustFromBig((*big.Int)(args.BlobFeeCap)),
		}
		if args.Blobs != nil {
			data.(*types.BlobTx).Sidecar = &types.BlobTxSidecar{
				Blobs:       args.Blobs,
				Commitments: args.Commitments,
				Proofs:      args.Proofs,
			}
		}

	case args.MaxFeePerGas != nil:
		al := types.AccessList{}
		if args.AccessList != nil {
			al = *args.AccessList
		}
		data = &types.DynamicFeeTx{
			To:         args.To,
			ChainID:    (*big.Int)(args.ChainID),
			Nonce:      uint64(*args.Nonce),
			Gas:        uint64(*args.Gas),
			GasFeeCap:  (*big.Int)(args.MaxFeePerGas),
			GasTipCap:  (*big.Int)(args.MaxPriorityFeePerGas),
			Value:      (*big.Int)(args.Value),
			Data:       args.data(),
			AccessList: al,
		}

	case args.AccessList != nil:
		data = &types.AccessListTx{
			To:         args.To,
			ChainID:    (*big.Int)(args.ChainID),
			Nonce:      uint64(*args.Nonce),
			Gas:        uint64(*args.Gas),
			GasPrice:   (*big.Int)(args.GasPrice),
			Value:      (*big.Int)(args.Value),
			Data:       args.data(),
			AccessList: *args.AccessList,
		}

	default:
		data = &types.LegacyTx{
			To:       args.To,
			Nonce:    uint64(*args.Nonce),
			Gas:      uint64(*args.Gas),
			GasPrice: (*big.Int)(args.GasPrice),
			Value:    (*big.Int)(args.Value),
			Data:     args.data(),
		}
	}
	return types.NewTx(data)
}

// IsEIP4844 returns an indicator if the args contains EIP4844 fields.
func (args *TransactionArgs) IsEIP4844() bool {
	return args.BlobHashes != nil || args.BlobFeeCap != nil
}
