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

// Package ethconfig contains the configuration of the ETH and LES protocols.
package ethconfig

import (
	"time"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/consensus"
	"github.com/ixios-io/ixiosSpark/consensus/fastClique"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/txpool/blobpool"
	"github.com/ixios-io/ixiosSpark/core/txpool/legacypool"
	"github.com/ixios-io/ixiosSpark/ixios/downloader"
	"github.com/ixios-io/ixiosSpark/ixios/gasprice"
	"github.com/ixios-io/ixiosSpark/kvdb"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/ixios-io/ixiosSpark/sealer"
)

// FullNodeGPO contains default gasprice oracle settings for full node.
var FullNodeGPO = gasprice.Config{
	Blocks:           20,
	Percentile:       60,
	MaxHeaderHistory: 1024,
	MaxBlockHistory:  1024,
	MaxPrice:         gasprice.DefaultMaxPrice,
	IgnorePrice:      gasprice.DefaultIgnorePrice,
}

// Defaults contains default settings for use on the Ixios main net.
var Defaults = Config{
	SyncMode:           downloader.FullSync,
	NetworkId:          0, // enable auto configuration of networkID == chainID
	TxLookupLimit:      31536000,
	TransactionHistory: 31536000,
	StateHistory:       params.FullImmutabilityThreshold,
	LightPeers:         0,
	DatabaseCache:      1024,
	TrieCleanCache:     256,
	TrieDirtyCache:     384,
	TrieTimeout:        60 * time.Minute,
	SnapshotCache:      0,
	FilterLogCacheSize: 32,
	Miner:              sealer.DefaultConfig,
	TxPool:             legacypool.DefaultConfig,
	BlobPool:           blobpool.DefaultConfig,
	RPCGasCap:          85287602,
	RPCEVMTimeout:      850 * time.Millisecond,
	GPO:                FullNodeGPO,
	RPCTxFeeCap:        42643801, // 42,643,801 IXO
	EnableBroadcast:    false,
}

//go:generate go run github.com/fjl/gencodec -type Config -formats toml -out gen_config.go

// Config contains configuration options for ETH and LES protocols.
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ixios main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Network ID separates blockchains on the peer-to-peer networking level. When left
	// zero, the chain ID is used as network ID.
	NetworkId uint64
	SyncMode  downloader.SyncMode

	// This can be set to list of enrtree:// URLs which will be queried for
	// for nodes to connect to.
	EthDiscoveryURLs  []string
	SnapDiscoveryURLs []string

	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

	// Deprecated, use 'TransactionHistory' instead.
	TxLookupLimit      uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	TransactionHistory uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	StateHistory       uint64 `toml:",omitempty"` // The maximum number of blocks from head whose state histories are reserved.

	// State scheme represents the scheme used to store ethereum states and trie
	// nodes on top. It can be 'hash', 'path', or none which means use the scheme
	// consistent with persistent state.
	StateScheme string `toml:",omitempty"`

	// RequiredBlocks is a set of block number -> hash mappings which must be in the
	// canonical chain of all remote peers. Setting the option makes ixiosSpark verify the
	// presence of these blocks for every new peer connection.
	RequiredBlocks map[uint64]common.Hash `toml:"-"`

	// Light client options
	LightServ        int  `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightIngress     int  `toml:",omitempty"` // Incoming bandwidth limit for light servers
	LightEgress      int  `toml:",omitempty"` // Outgoing bandwidth limit for light servers
	LightPeers       int  `toml:",omitempty"` // Maximum number of LES client peers
	LightNoPrune     bool `toml:",omitempty"` // Whether to disable light chain pruning
	LightNoSyncServe bool `toml:",omitempty"` // Whether to serve light clients before syncing

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration
	SnapshotCache  int
	Preimages      bool

	// This is the number of blocks for which logs will be cached in the filter system.
	FilterLogCacheSize int

	// Mining options
	Miner sealer.Config

	// Transaction pool options
	TxPool   legacypool.Config
	BlobPool blobpool.Config

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap uint64

	// RPCEVMTimeout is the global timeout for eth-call.
	RPCEVMTimeout time.Duration

	// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	// send-transaction variants. The unit is ether.
	RPCTxFeeCap float64

	EnableBroadcast bool
}

// CreateConsensusEngine creates a consensus engine for the given chain config.
func CreateConsensusEngine(config *params.ChainConfig, db kvdb.Database) (consensus.Engine, error) {
	if config.Clique != nil {
		return fastClique.New(config.Clique, db), nil
	}

	panic("Unsupported consensus engine")
}
