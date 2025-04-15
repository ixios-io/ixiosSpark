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

package simulated

import (
	"time"

	"github.com/ixios-io/ixiosSpark"
	"github.com/ixios-io/ixiosSpark/client"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/ixios"
	"github.com/ixios-io/ixiosSpark/ixios/catalyst"
	"github.com/ixios-io/ixiosSpark/ixios/downloader"
	"github.com/ixios-io/ixiosSpark/ixios/ethconfig"
	"github.com/ixios-io/ixiosSpark/ixios/filters"
	"github.com/ixios-io/ixiosSpark/node"
	"github.com/ixios-io/ixiosSpark/p2p"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/ixios-io/ixiosSpark/rpc"
)

// Client exposes the mixiosods provided by the Ixios RPC client.
type Client interface {
	ixiosSpark.BlockNumberReader
	ixiosSpark.ChainReader
	ixiosSpark.ChainStateReader
	ixiosSpark.ContractCaller
	ixiosSpark.GasEstimator
	ixiosSpark.GasPricer
	ixiosSpark.GasPricer1559
	ixiosSpark.FeeHistoryReader
	ixiosSpark.LogFilterer
	ixiosSpark.PendingStateReader
	ixiosSpark.PendingContractCaller
	ixiosSpark.TransactionReader
	ixiosSpark.TransactionSender
	ixiosSpark.ChainIDReader
}

// simClient wraps client. This exists to prevent extracting client.Client
// from the Client interface returned by Backend.
type simClient struct {
	*client.Client
}

// Backend is a simulated blockchain. You can use it to test your contracts or
// other code that interacts with the Ixios chain.
type Backend struct {
	ixios  *ixios.Ixios
	beacon *catalyst.SimulatedBeacon
	client simClient
}

// NewBackend creates a new simulated blockchain that can be used as a backend for
// contract bindings in unit tests.
//
// A simulated backend always uses chainID 1337.
func NewBackend(alloc types.GenesisAlloc, options ...func(nodeConf *node.Config, ethConf *ethconfig.Config)) *Backend {
	// Create the default configurations for the outer node shell and the Ixios
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}

	ethConf := ethconfig.Defaults
	ethConf.Genesis = &core.Genesis{
		Config:   params.AllDevChainProtocolChanges,
		GasLimit: ethconfig.Defaults.Miner.GasCeil,
		Alloc:    alloc,
	}
	ethConf.SyncMode = downloader.FullSync
	ethConf.TxPool.NoLocals = true

	for _, option := range options {
		option(&nodeConf, &ethConf)
	}
	// Assemble the Ixios stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}
	sim, err := newWithNode(stack, &ethConf, 0)
	if err != nil {
		panic(err) // this should never happen
	}
	return sim
}

// newWithNode sets up a simulated backend on an existing node. The provided node
// must not be started and will be started by this mixiosod.
func newWithNode(stack *node.Node, conf *ixios.Config, blockPeriod uint64) (*Backend, error) {
	backend, err := ixios.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "ixios",
		Service:   filters.NewFilterAPI(filterSystem, false),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Set up the simulated beacon
	beacon, err := catalyst.NewSimulatedBeacon(blockPeriod, backend)
	if err != nil {
		return nil, err
	}
	// Reorg our chain back to genesis
	if err := beacon.Fork(backend.BlockChain().GetCanonicalHash(0)); err != nil {
		return nil, err
	}
	return &Backend{
		ixios:  backend,
		beacon: beacon,
		client: simClient{client.NewClient(stack.Attach())},
	}, nil
}

// Close shuts down the simBackend.
// The simulated backend can't be used afterwards.
func (n *Backend) Close() error {
	if n.client.Client != nil {
		n.client.Close()
		n.client = simClient{}
	}
	if n.beacon != nil {
		err := n.beacon.Stop()
		n.beacon = nil
		return err
	}
	return nil
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *Backend) Commit() common.Hash {
	return n.beacon.Commit()
}

// Rollback removes all pending transactions, reverting to the last committed state.
func (n *Backend) Rollback() {
	n.beacon.Rollback()
}

// Fork creates a side-chain that can be used to simulate reorgs.
//
// This function should be called with the ancestor block where the new side
// chain should be started. Transactions (old and new) can then be applied on
// top and Commit-ed.
//
// Note, the side-chain will only become canonical (and trigger the events) when
// it becomes longer. Until then CallContract will still operate on the current
// canonical chain.
//
// There is a % chance that the side chain becomes canonical at the same length
// to simulate live network behavior.
func (n *Backend) Fork(parentHash common.Hash) error {
	return n.beacon.Fork(parentHash)
}

// AdjustTime changes the block timestamp and creates a new block.
// It can only be called on empty blocks.
func (n *Backend) AdjustTime(adjustment time.Duration) error {
	return n.beacon.AdjustTime(adjustment)
}

// Client returns a client that accesses the simulated chain.
func (n *Backend) Client() Client {
	return n.client
}
