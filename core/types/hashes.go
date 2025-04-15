// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors, Copyright 2015-2024 The go-ethereum Authors (geth)
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/crypto"
	"github.com/ixios-io/ixiosSpark/log"
)

var (
	// EmptyRootHash is the known root hash of an empty merkle trie.
	EmptyRootHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// EmptyOmmerHash is the known hash of the empty uncle set.
	EmptyOmmerHash = rlpHash([]*Header(nil)) // 1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347

	// EmptyCodeHash is the known hash of the empty EVM bytecode.
	EmptyCodeHash = crypto.Keccak256Hash(nil) // c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470

	// EmptyTxsHash is the known hash of the empty transaction set.
	EmptyTxsHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// EmptyReceiptsHash is the known hash of the empty receipt set.
	EmptyReceiptsHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// EmptyWithdrawalsHash is the known hash of the empty withdrawal set.
	EmptyWithdrawalsHash = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// EmptyVerkleHash is the known hash of an empty verkle trie.
	EmptyVerkleHash = common.Hash{}
)

// TrieRootHash returns the hash itself if it's non-empty or the predefined
// emptyHash one instead.
func TrieRootHash(hash common.Hash) common.Hash {
	if hash == (common.Hash{}) {
		log.Error("Zero trie root hash!")
		return EmptyRootHash
	}
	return hash
}
