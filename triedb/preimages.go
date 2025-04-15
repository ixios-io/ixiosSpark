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

package triedb

import (
	"sync"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/rawdb"
	"github.com/ixios-io/ixiosSpark/kvdb"
)

// preimageStore is the store for caching preimages of node key.
type preimageStore struct {
	lock          sync.RWMutex
	disk          kvdb.KeyValueStore
	preimages     map[common.Hash][]byte // Preimages of nodes from the secure trie
	preimagesSize common.StorageSize     // Storage size of the preimages cache
}

// newPreimageStore initializes the store for caching preimages.
func newPreimageStore(disk kvdb.KeyValueStore) *preimageStore {
	return &preimageStore{
		disk:      disk,
		preimages: make(map[common.Hash][]byte),
	}
}

// insertPreimage writes a new trie node pre-image to the memory database if it's
// yet unknown. The method will NOT make a copy of the slice, only use if the
// preimage will NOT be changed later on.
func (store *preimageStore) insertPreimage(preimages map[common.Hash][]byte) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for hash, preimage := range preimages {
		if _, ok := store.preimages[hash]; ok {
			continue
		}
		store.preimages[hash] = preimage
		store.preimagesSize += common.StorageSize(common.HashLength + len(preimage))
	}
}

// preimage retrieves a cached trie node pre-image from memory. If it cannot be
// found cached, the method queries the persistent database for the content.
func (store *preimageStore) preimage(hash common.Hash) []byte {
	store.lock.RLock()
	preimage := store.preimages[hash]
	store.lock.RUnlock()

	if preimage != nil {
		return preimage
	}
	return rawdb.ReadPreimage(store.disk, hash)
}

// commit flushes the cached preimages into the disk.
func (store *preimageStore) commit(force bool) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	if store.preimagesSize <= 4*1024*1024 && !force {
		return nil
	}
	batch := store.disk.NewBatch()
	rawdb.WritePreimages(batch, store.preimages)
	if err := batch.Write(); err != nil {
		return err
	}
	store.preimages, store.preimagesSize = make(map[common.Hash][]byte), 0
	return nil
}

// size returns the current storage size of accumulated preimages.
func (store *preimageStore) size() common.StorageSize {
	store.lock.RLock()
	defer store.lock.RUnlock()

	return store.preimagesSize
}
