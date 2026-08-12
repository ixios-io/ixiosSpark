// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/params"
)

func isAddressFormatActive(config *params.ChainConfig, blockNumber *big.Int) bool {
	return config != nil && config.IsAddressFormat(blockNumber)
}

func isAddressAllowedAtBlock(config *params.ChainConfig, blockNumber *big.Int, addr common.Address) bool {
	if isAddressFormatActive(config, blockNumber) {
		return true
	}
	return addr.IsLegacyCompatible()
}

// ValidateTxAddressFormats enforces the Aegis Upgrade fork. Before the fork,
// transactions must remain fully legacy-compatible: sender signatures must be
// ECDSA and every address carried by the transaction must fit the historic
// 20/32-byte address encoding. After the fork, both legacy-compatible and full
// 48-byte addresses are valid.
func ValidateTxAddressFormats(config *params.ChainConfig, blockNumber *big.Int, tx *types.Transaction, from common.Address) error {
	if !isAddressFormatActive(config, blockNumber) && bytes.Equal(tx.SignatureType(), types.SigTypeMLDSA87) {
		return fmt.Errorf("MLDSA87 signature payload not allowed before Aegis Upgrade hard fork at block %v", blockNumber)
	}
	if !isAddressAllowedAtBlock(config, blockNumber, from) {
		return fmt.Errorf("non-legacy sender address not allowed before Aegis Upgrade hard fork at block %v", blockNumber)
	}
	if to := tx.To(); to != nil && !isAddressAllowedAtBlock(config, blockNumber, *to) {
		return fmt.Errorf("non-legacy recipient address not allowed before Aegis Upgrade hard fork at block %v", blockNumber)
	}
	for i, tuple := range tx.AccessList() {
		if !isAddressAllowedAtBlock(config, blockNumber, tuple.Address) {
			return fmt.Errorf("non-legacy access list entry %d not allowed before Aegis Upgrade hard fork at block %v", i, blockNumber)
		}
	}
	return nil
}
