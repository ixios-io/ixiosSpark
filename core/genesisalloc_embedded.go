// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// Copyright 2025 The IxiosSpark Authors
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/types"
)

//go:embed data/genesisalloc.dat
var embeddedGenesisAlloc []byte

// ParseGenesisAllocData reads the in-memory bytes from embeddedGenesisAlloc,
// in 64-byte blocks (32 bytes for the address, 32 bytes for the balance).
// Returns a map[common.Address]types.Account
func ParseGenesisAllocData() (map[common.Address]types.Account, error) {
	const (
		chunkSize   = 64
		addressSize = 32
	)

	dataLen := len(embeddedGenesisAlloc)
	if dataLen%chunkSize != 0 {
		return nil, fmt.Errorf("embedded genesisalloc.dat size %d is not multiple of %d", dataLen, chunkSize)
	}

	allocs := make(map[common.Address]types.Account)

	for i := 0; i < dataLen; i += chunkSize {
		chunk := embeddedGenesisAlloc[i : i+chunkSize]

		// address
		var addr common.Address
		copy(addr[:], chunk[:addressSize])

		// balance
		balanceBytes := chunk[addressSize:chunkSize]
		balance := new(big.Int).SetBytes(balanceBytes)

		allocs[addr] = types.Account{
			Balance: balance,
		}
	}
	return allocs, nil
}
