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

// Package sealer implements block creation and sealing.
package sealer

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
	"github.com/ixios-io/ixiosSpark/consensus"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/state"
	"github.com/ixios-io/ixiosSpark/core/txpool"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/event"
	"github.com/ixios-io/ixiosSpark/ixios/downloader"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/params"
)

// Backend wraps all methods required for mining. Only full node is capable
// to offer all the functions here.
type Backend interface {
	BlockChain() *core.BlockChain
	TxPool() *txpool.TxPool
}

// Config is the configuration parameters of mining.
type Config struct {
	Etherbase common.Address `toml:",omitempty"` // Public address for block mining rewards
	ExtraData hexutil.Bytes  `toml:",omitempty"` // Block extra data set by the sealer
	GasFloor  uint64         // Target gas floor for mined blocks.
	GasCeil   uint64         // Target gas ceiling for mined blocks.
	GasPrice  *big.Int       // Minimum gas price for mining a transaction
	Recommit  time.Duration  // The time interval for sealer to re-create mining work.

	NewPayloadTimeout time.Duration // The maximum time allowance for creating a new payload
}

// DefaultConfig contains default settings for sealer.
var DefaultConfig = Config{
	GasCeil:  30000000,
	GasPrice: big.NewInt(params.GWei),

	// For Ixios with 1s blocks:
	// - Give ~350ms for tx execution and block building
	// - Allow time for network propagation
	// - Leave buffer for next block
	Recommit:          350 * time.Millisecond,
	NewPayloadTimeout: 350 * time.Millisecond,
}

// Sealer creates blocks
type Sealer struct {
	mux     *event.TypeMux
	eth     Backend
	engine  consensus.Engine
	exitCh  chan struct{}
	startCh chan struct{}
	stopCh  chan struct{}
	worker  *worker

	wg sync.WaitGroup
}

func New(eth Backend, config *Config, chainConfig *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine, isLocalBlock func(header *types.Header) bool) *Sealer {
	sealer := &Sealer{
		mux:     mux,
		eth:     eth,
		engine:  engine,
		exitCh:  make(chan struct{}),
		startCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
		worker:  newWorker(config, chainConfig, engine, eth, mux, isLocalBlock, true),
	}
	sealer.wg.Add(1)
	go sealer.update()
	return sealer
}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your mining operation for as long as the DOS continues.
func (sealer *Sealer) update() {
	defer sealer.wg.Done()

	events := sealer.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
	defer func() {
		if !events.Closed() {
			events.Unsubscribe()
		}
	}()

	shouldStart := false
	canStart := true
	dlEventCh := events.Chan()
	for {
		select {
		case ev := <-dlEventCh:
			if ev == nil {
				// Unsubscription done, stop listening
				dlEventCh = nil
				continue
			}
			switch ev.Data.(type) {
			case downloader.StartEvent:
				wasMining := sealer.Mining()
				sealer.worker.stop()
				canStart = false
				if wasMining {
					// Resume mining after sync was finished
					shouldStart = true
					log.Info("Sealing aborted due to sync")
				}
				sealer.worker.syncing.Store(true)

			case downloader.FailedEvent:
				canStart = true
				if shouldStart {
					sealer.worker.start()
				}
				sealer.worker.syncing.Store(false)

			case downloader.DoneEvent:
				canStart = true
				if shouldStart {
					sealer.worker.start()
				}
				sealer.worker.syncing.Store(false)

				// Stop reacting to downloader events
				events.Unsubscribe()
			}
		case <-sealer.startCh:
			if canStart {
				// Add 3 second delay on initial start
				log.Info("Waiting ~3 seconds before starting sealer")
				time.Sleep(2500 * time.Millisecond)

				// Add some randomness so sealers don't start at the same time
				time.Sleep(1000 * time.Millisecond)
				sealer.worker.start()
			}
			shouldStart = true
		case <-sealer.stopCh:
			shouldStart = false
			sealer.worker.stop()
		case <-sealer.exitCh:
			sealer.worker.close()
			return
		}
	}
}

func (sealer *Sealer) Start() {
	sealer.startCh <- struct{}{}
}

func (sealer *Sealer) Stop() {
	sealer.stopCh <- struct{}{}
}

func (sealer *Sealer) Close() {
	close(sealer.exitCh)
	sealer.wg.Wait()
}

func (sealer *Sealer) Mining() bool {
	return sealer.worker.isRunning()
}

func (sealer *Sealer) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	sealer.worker.setExtra(extra)
	return nil
}

func (sealer *Sealer) SetGasTip(tip *big.Int) error {
	sealer.worker.setGasTip(tip)
	return nil
}

// SetRecommitInterval sets the interval for sealing work resubmitting.
func (sealer *Sealer) SetRecommitInterval(interval time.Duration) {
	sealer.worker.setRecommitInterval(interval)
}

// Pending returns the currently pending block and associated state. The returned
// values can be nil in case the pending block is not initialized
func (sealer *Sealer) Pending() (*types.Block, *state.StateDB) {
	return sealer.worker.pending()
}

// PendingBlock returns the currently pending block. The returned block can be
// nil in case the pending block is not initialized.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
func (sealer *Sealer) PendingBlock() *types.Block {
	return sealer.worker.pendingBlock()
}

// PendingBlockAndReceipts returns the currently pending block and corresponding receipts.
// The returned values can be nil in case the pending block is not initialized.
func (sealer *Sealer) PendingBlockAndReceipts() (*types.Block, types.Receipts) {
	return sealer.worker.pendingBlockAndReceipts()
}

func (sealer *Sealer) SetEtherbase(addr common.Address) {
	sealer.worker.setEtherbase(addr)
}

// SetGasCeil sets the gaslimit to strive for when mining blocks post 1559.
// For pre-1559 blocks, it sets the ceiling.
func (sealer *Sealer) SetGasCeil(ceil uint64) {
	sealer.worker.setGasCeil(ceil)
}

// SubscribePendingLogs starts delivering logs from pending transactions
// to the given channel.
func (sealer *Sealer) SubscribePendingLogs(ch chan<- []*types.Log) event.Subscription {
	return sealer.worker.pendingLogsFeed.Subscribe(ch)
}

// BuildPayload builds the payload according to the provided parameters.
func (sealer *Sealer) BuildPayload(args *BuildPayloadArgs) (*Payload, error) {
	return sealer.worker.buildPayload(args)
}
