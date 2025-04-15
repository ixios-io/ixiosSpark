// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2015-2024 The go-ethereum Authors (geth)
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package trie

import "github.com/ixios-io/ixiosSpark/common"

// ID is the identifier for uniquely identifying a trie.
type ID struct {
	StateRoot common.Hash // The root of the corresponding state(block.root)
	Owner     common.Hash // The contract address hash which the trie belongs to
	Root      common.Hash // The root hash of trie
}

// StateTrieID constructs an identifier for state trie with the provided state root.
func StateTrieID(root common.Hash) *ID {
	return &ID{
		StateRoot: root,
		Owner:     common.Hash{},
		Root:      root,
	}
}

// StorageTrieID constructs an identifier for storage trie which belongs to a certain
// state and contract specified by the stateRoot and owner.
func StorageTrieID(stateRoot common.Hash, owner common.Hash, root common.Hash) *ID {
	return &ID{
		StateRoot: stateRoot,
		Owner:     owner,
		Root:      root,
	}
}

// TrieID constructs an identifier for a standard trie(not a second-layer trie)
// with provided root. It's mostly used in tests and some other tries like CHT trie.
func TrieID(root common.Hash) *ID {
	return &ID{
		StateRoot: root,
		Owner:     common.Hash{},
		Root:      root,
	}
}
