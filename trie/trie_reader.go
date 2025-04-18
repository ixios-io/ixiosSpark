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

import (
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/trie/triestate"
	"github.com/ixios-io/ixiosSpark/triedb/database"
)

// trieReader is a wrapper of the underlying node reader. It's not safe
// for concurrent usage.
type trieReader struct {
	owner  common.Hash
	reader database.Reader
	banned map[string]struct{} // Marker to prevent node from being accessed, for tests
}

// newTrieReader initializes the trie reader with the given node reader.
func newTrieReader(stateRoot, owner common.Hash, db database.Database) (*trieReader, error) {
	if stateRoot == (common.Hash{}) || stateRoot == types.EmptyRootHash {
		if stateRoot == (common.Hash{}) {
			log.Error("Zero state root hash!")
		}
		return &trieReader{owner: owner}, nil
	}
	reader, err := db.Reader(stateRoot)
	if err != nil {
		return nil, &MissingNodeError{Owner: owner, NodeHash: stateRoot, err: err}
	}
	return &trieReader{owner: owner, reader: reader}, nil
}

// newEmptyReader initializes the pure in-memory reader. All read operations
// should be forbidden and returns the MissingNodeError.
func newEmptyReader() *trieReader {
	return &trieReader{}
}

// node retrieves the rlp-encoded trie node with the provided trie node
// information. An MissingNodeError will be returned in case the node is
// not found or any error is encountered.
func (r *trieReader) node(path []byte, hash common.Hash) ([]byte, error) {
	// Perform the logics in tests for preventing trie node access.
	if r.banned != nil {
		if _, ok := r.banned[string(path)]; ok {
			return nil, &MissingNodeError{Owner: r.owner, NodeHash: hash, Path: path}
		}
	}
	if r.reader == nil {
		return nil, &MissingNodeError{Owner: r.owner, NodeHash: hash, Path: path}
	}
	blob, err := r.reader.Node(r.owner, path, hash)
	if err != nil || len(blob) == 0 {
		return nil, &MissingNodeError{Owner: r.owner, NodeHash: hash, Path: path, err: err}
	}
	return blob, nil
}

// MerkleLoader implements triestate.TrieLoader for constructing tries.
type MerkleLoader struct {
	db database.Database
}

// NewMerkleLoader creates the merkle trie loader.
func NewMerkleLoader(db database.Database) *MerkleLoader {
	return &MerkleLoader{db: db}
}

// OpenTrie opens the main account trie.
func (l *MerkleLoader) OpenTrie(root common.Hash) (triestate.Trie, error) {
	return New(TrieID(root), l.db)
}

// OpenStorageTrie opens the storage trie of an account.
func (l *MerkleLoader) OpenStorageTrie(stateRoot common.Hash, addrHash, root common.Hash) (triestate.Trie, error) {
	return New(StorageTrieID(stateRoot, addrHash, root), l.db)
}
