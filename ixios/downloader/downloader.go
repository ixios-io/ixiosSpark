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

// Package downloader contains the manual full chain synchronisation.
package downloader

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ixios-io/ixiosSpark"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/rawdb"
	"github.com/ixios-io/ixiosSpark/core/state/snapshot"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/event"
	"github.com/ixios-io/ixiosSpark/kvdb"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/ixios-io/ixiosSpark/triedb"
)

var (
	MaxBlockFetch   = 256 // Amount of blocks to be fetched per retrieval request
	MaxHeaderFetch  = 768 // Amount of block headers to be fetched per retrieval request
	MaxSkeletonSize = 192 // Number of header fetches to need for a skeleton assembly
	MaxReceiptFetch = 384 // Amount of transaction receipts to allow fetching per request

	maxQueuedHeaders            = 32 * 1024                         // Maximum number of headers to queue for import (DOS protection)
	maxHeadersProcess           = 2048                              // Number of header download results to import at once into the chain
	maxResultsProcess           = 2048                              // Number of content download results to import at once into the chain
	fullMaxForkAncestry  uint64 = params.FullImmutabilityThreshold  // Maximum chain reorganisation (locally redeclared so tests can reduce it)
	lightMaxForkAncestry uint64 = params.LightImmutabilityThreshold // Maximum chain reorganisation (locally redeclared so tests can reduce it)

	reorgProtThreshold   = 64 // Threshold number of recent blocks to disable mini reorg protection
	reorgProtHeaderDelay = 2  // Number of headers to delay delivering to cover mini reorgs

	fsHeaderSafetyNet = 2048                   // Number of headers to discard in case a chain violation is detected
	fsHeaderContCheck = 300 * time.Millisecond // Time interval to check for header continuations during state download
	fsMinFullBlocks   = 64                     // Number of blocks to retrieve fully even in snap sync
)

var (
	errBusy                    = errors.New("busy")
	errUnknownPeer             = errors.New("peer is unknown or unhealthy")
	errBadPeer                 = errors.New("action from bad peer ignored")
	errStallingPeer            = errors.New("peer is stalling")
	errUnsyncedPeer            = errors.New("unsynced peer")
	errNoPeers                 = errors.New("no peers to keep download active")
	errTimeout                 = errors.New("timeout")
	errEmptyHeaderSet          = errors.New("empty header set by peer")
	errPeersUnavailable        = errors.New("no peers available or all tried for download")
	errInvalidAncestor         = errors.New("retrieved ancestor is invalid")
	errInvalidChain            = errors.New("retrieved hash chain is invalid")
	errInvalidBody             = errors.New("retrieved block body is invalid")
	errInvalidReceipt          = errors.New("retrieved receipt is invalid")
	errCancelStateFetch        = errors.New("state data download canceled (requested)")
	errCancelContentProcessing = errors.New("content processing canceled (requested)")
	errCanceled                = errors.New("syncing canceled (requested)")
	errTooOld                  = errors.New("peer's protocol version too old")
	errNoAncestorFound         = errors.New("no common ancestor found")
	errNoPivotHeader           = errors.New("pivot header is not found")
	ErrMergeTransition         = errors.New("legacy sync reached the merge")
)

// peerDropFn is a callback type for dropping a peer detected as malicious.
type peerDropFn func(id string)

// badBlockFn is a callback for the async beacon sync to notify the caller that
// the origin header requested to sync to, produced a chain with a bad block.
type badBlockFn func(invalid *types.Header, origin *types.Header)

// headerTask is a set of downloaded headers to queue along with their precomputed
// hashes to avoid constant rehashing.
type headerTask struct {
	headers []*types.Header
	hashes  []common.Hash
}

type Downloader struct {
	mode atomic.Uint32  // Synchronisation mode defining the strategy used (per sync cycle), use d.getMode() to get the SyncMode
	mux  *event.TypeMux // Event multiplexer to announce sync operation events

	genesis uint64   // Genesis block number to limit sync to (e.g. light client CHT)
	queue   *queue   // Scheduler for selecting the hashes to download
	peers   *peerSet // Set of active peers from which download can proceed

	stateDB kvdb.Database // Database to state sync into (and deduplicate via)

	// Statistics
	syncStatsChainOrigin uint64       // Origin block number where syncing started at
	syncStatsChainHeight uint64       // Highest block number known when syncing started
	syncStatsLock        sync.RWMutex // Lock protecting the sync stats fields

	lightchain LightChain
	blockchain BlockChain

	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving
	badBlock badBlockFn // Reports a block as rejected by the chain

	// Status
	synchroniseMock func(id string, hash common.Hash) error // Replacement for synchronise during testing
	synchronising   atomic.Bool
	notified        atomic.Bool
	committed       atomic.Bool
	ancientLimit    uint64 // The maximum block number which can be regarded as ancient data.

	// Channels
	headerProcCh chan *headerTask // Channel to feed the header processor new tasks

	// Skeleton sync
	skeleton *skeleton // Header skeleton to backfill the chain with (eth2 mode)

	// State sync
	pivotHeader    *types.Header // Pivot block header to dynamically push the syncing state root
	pivotLock      sync.RWMutex  // Lock protecting pivot header reads from updates
	stateSyncStart chan *stateSync

	// Cancellation and termination
	cancelPeer string         // Identifier of the peer currently being used as the master (cancel on drop)
	cancelCh   chan struct{}  // Channel to cancel mid-flight syncs
	cancelLock sync.RWMutex   // Lock to protect the cancel channel and peer in delivers
	cancelWg   sync.WaitGroup // Make sure all fetcher goroutines have exited.

	quitCh   chan struct{} // Quit channel to signal termination
	quitLock sync.Mutex    // Lock to prevent double closes

	// Testing hooks
	syncInitHook     func(uint64, uint64)  // Method to call upon initiating a new sync run
	bodyFetchHook    func([]*types.Header) // Method to call upon starting a block body fetch
	receiptFetchHook func([]*types.Header) // Method to call upon starting a receipt fetch
	chainInsertHook  func([]*fetchResult)  // Method to call upon inserting a chain of blocks (possibly in multiple invocations)

	// Progress reporting metrics
	syncStartBlock uint64    // Head snap block when Geth was started
	syncStartTime  time.Time // Time instance when chain sync started
	syncLogTime    time.Time // Time instance when status was last reported
}

// LightChain encapsulates functions required to synchronise a light chain.
type LightChain interface {
	// HasHeader verifies a header's presence in the local chain.
	HasHeader(common.Hash, uint64) bool

	// GetHeaderByHash retrieves a header from the local chain.
	GetHeaderByHash(common.Hash) *types.Header

	// CurrentHeader retrieves the head header from the local chain.
	CurrentHeader() *types.Header

	// GetTd returns the total difficulty of a local block.
	GetTd(common.Hash, uint64) *big.Int

	// InsertHeaderChain inserts a batch of headers into the local chain.
	InsertHeaderChain([]*types.Header) (int, error)

	// SetHead rewinds the local chain to a new head.
	SetHead(uint64) error
}

// BlockChain encapsulates functions required to sync a (full or snap) blockchain.
type BlockChain interface {
	LightChain

	// HasBlock verifies a block's presence in the local chain.
	HasBlock(common.Hash, uint64) bool

	// HasFastBlock verifies a snap block's presence in the local chain.
	HasFastBlock(common.Hash, uint64) bool

	// GetBlockByHash retrieves a block from the local chain.
	GetBlockByHash(common.Hash) *types.Block

	// CurrentBlock retrieves the head block from the local chain.
	CurrentBlock() *types.Header

	// CurrentSnapBlock retrieves the head snap block from the local chain.
	CurrentSnapBlock() *types.Header

	// SnapSyncCommitHead directly commits the head block to a certain entity.
	SnapSyncCommitHead(common.Hash) error

	// InsertChain inserts a batch of blocks into the local chain.
	InsertChain(types.Blocks) (int, error)

	// InsertReceiptChain inserts a batch of receipts into the local chain.
	InsertReceiptChain(types.Blocks, []types.Receipts, uint64) (int, error)

	// Snapshots returns the blockchain snapshot tree to paused it during sync.
	Snapshots() *snapshot.Tree

	// TrieDB retrieves the low level trie database used for interacting
	// with trie nodes.
	TrieDB() *triedb.Database
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(stateDb kvdb.Database, mux *event.TypeMux, chain BlockChain, lightchain LightChain, dropPeer peerDropFn, success func()) *Downloader {
	if lightchain == nil {
		lightchain = chain
	}
	dl := &Downloader{
		stateDB:        stateDb,
		mux:            mux,
		queue:          newQueue(blockCacheMaxItems, blockCacheInitialItems),
		peers:          newPeerSet(),
		blockchain:     chain,
		lightchain:     lightchain,
		dropPeer:       dropPeer,
		headerProcCh:   make(chan *headerTask, 1),
		quitCh:         make(chan struct{}),
		stateSyncStart: make(chan *stateSync),
		syncStartBlock: chain.CurrentSnapBlock().Number.Uint64(),
	}
	// Create the post-merge skeleton syncer and start the process
	dl.skeleton = newSkeleton(stateDb, dl.peers, dropPeer, newBeaconBackfiller(dl, success))

	go dl.stateFetcher()
	return dl
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of snap synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
func (d *Downloader) Progress() ixiosSpark.SyncProgress {
	// Lock the current stats and return the progress
	d.syncStatsLock.RLock()
	defer d.syncStatsLock.RUnlock()

	current := uint64(0)
	mode := d.getMode()
	switch {
	case d.blockchain != nil && mode == FullSync:
		current = d.blockchain.CurrentBlock().Number.Uint64()
	case d.blockchain != nil && mode == SnapSync:
		current = d.blockchain.CurrentSnapBlock().Number.Uint64()
	case d.lightchain != nil:
		current = d.lightchain.CurrentHeader().Number.Uint64()
	default:
		log.Error("Unknown downloader chain/mode combo", "light", d.lightchain != nil, "full", d.blockchain != nil, "mode", mode)
	}

	return ixiosSpark.SyncProgress{
		StartingBlock: d.syncStatsChainOrigin,
		CurrentBlock:  current,
		HighestBlock:  d.syncStatsChainHeight,
	}
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (d *Downloader) RegisterPeer(id string, version uint, peer Peer) error {
	var logger log.Logger
	if len(id) < 16 {
		// Tests use short IDs, don't choke on them
		logger = log.New("peer", id)
	} else {
		logger = log.New("peer", id[:8])
	}
	logger.Trace("Registering sync peer")
	if err := d.peers.Register(newPeerConnection(id, version, peer, logger)); err != nil {
		logger.Error("Failed to register sync peer", "err", err)
		return err
	}
	return nil
}

// UnregisterPeer remove a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (d *Downloader) UnregisterPeer(id string) error {
	// Unregister the peer from the active peer set and revoke any fetch tasks
	var logger log.Logger
	if len(id) < 16 {
		// Tests use short IDs, don't choke on them
		logger = log.New("peer", id)
	} else {
		logger = log.New("peer", id[:8])
	}
	logger.Trace("Unregistering sync peer")
	if err := d.peers.Unregister(id); err != nil {
		logger.Error("Failed to unregister sync peer", "err", err)
		return err
	}
	d.queue.Revoke(id)

	return nil
}

// LegacySync tries to sync up our local blockchain with a remote peer, both
// adding various sanity checks, as well as wrapping it with various log entries.
func (d *Downloader) LegacySync(id string, head common.Hash, td, ttd *big.Int, mode SyncMode) error {
	err := d.synchronise(id, head, td, ttd, mode, false, nil)

	switch err {
	case nil, errBusy, errCanceled:
		return err
	}
	if errors.Is(err, errInvalidChain) || errors.Is(err, errBadPeer) || errors.Is(err, errTimeout) ||
		errors.Is(err, errStallingPeer) || errors.Is(err, errUnsyncedPeer) || errors.Is(err, errEmptyHeaderSet) ||
		errors.Is(err, errPeersUnavailable) || errors.Is(err, errTooOld) || errors.Is(err, errInvalidAncestor) {
		log.Warn("Synchronisation failed, dropping peer", "peer", id, "err", err)
		if d.dropPeer == nil {
			// The dropPeer method is nil when `--copydb` is used for a local copy.
			// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
			log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", id)
		} else {
			d.dropPeer(id)
		}
		return err
	}
	if errors.Is(err, ErrMergeTransition) {
		return err // This is an expected fault, don't keep printing it in a spin-loop
	}
	log.Warn("Synchronisation failed, retrying", "err", err)
	return err
}

// synchronise will select the peer and use it for synchronising. If an empty string is given
// it will use the best peer possible and synchronise if its TD is higher than our own. If any of the
// checks fail an error will be returned. This method is synchronous
func (d *Downloader) synchronise(id string, hash common.Hash, td, ttd *big.Int, mode SyncMode, beaconMode bool, beaconPing chan struct{}) error {
	// Make sure only one goroutine is ever allowed past this point at once
	if !d.synchronising.CompareAndSwap(false, true) {
		return errBusy
	}
	defer d.synchronising.Store(false)

	// Post a user notification of the sync (only once per session)
	if d.notified.CompareAndSwap(false, true) {
		log.Info("Block synchronisation started")
	}
	// Reset the queue, peer set and wake channels to clean any internal leftover state
	d.queue.Reset(blockCacheMaxItems, blockCacheInitialItems)
	d.peers.Reset()

	for _, ch := range []chan bool{d.queue.blockWakeCh, d.queue.receiptWakeCh} {
		select {
		case <-ch:
		default:
		}
	}
	for empty := false; !empty; {
		select {
		case <-d.headerProcCh:
		default:
			empty = true
		}
	}
	// Create cancel channel for aborting mid-flight and mark the master peer
	defer d.Cancel() // No matter what, we can't leave the cancel channel open

	if !d.cancelLock.TryLock() {
		return errors.New("Failed to acquire cancelLock (downloader.go) in synchronise()")
	}
	d.cancelCh = make(chan struct{})
	d.cancelPeer = id
	d.cancelLock.Unlock()

	// Atomically set the requested sync mode
	d.mode.Store(uint32(mode))

	// Retrieve the origin peer and initiate the downloading process
	var p *peerConnection
	p = d.peers.Peer(id)
	if p == nil {
		return errUnknownPeer
	}

	return d.syncWithPeer(p, hash, td, ttd, beaconMode)
}

func (d *Downloader) getMode() SyncMode {
	return SyncMode(d.mode.Load())
}

// syncWithPeer starts a block synchronization based on the hash chain from the
// specified peer and head hash.B
func (d *Downloader) syncWithPeer(p *peerConnection, hash common.Hash, td, ttd *big.Int, beaconMode bool) (err error) {
	d.mux.Post(StartEvent{})
	defer func() {
		// reset on error
		if err != nil {
			d.mux.Post(FailedEvent{err})
		} else {
			latest := d.lightchain.CurrentHeader()
			d.mux.Post(DoneEvent{latest})
		}
	}()
	mode := d.getMode()

	log.Debug("Synchronising with the network", "peer", p.id, "ixios", p.version, "head", hash, "td", td, "mode", mode)
	defer func(start time.Time) {
		log.Debug("Synchronisation stopped", "elapsed", common.PrettyDuration(time.Since(start)))
	}(time.Now())

	// Look up the sync boundaries: the common ancestor and the target block
	var latest *types.Header
	latest, _, err = d.fetchHead(p)
	if err != nil {
		return err
	}
	height := latest.Number.Uint64()

	var origin uint64

	// reach out to the network and find the ancestor
	origin, err = d.findAncestor(p, latest)
	if err != nil {
		return err
	}

	d.syncStatsLock.Lock()
	if d.syncStatsChainHeight <= origin || d.syncStatsChainOrigin > origin {
		d.syncStatsChainOrigin = origin
	}
	d.syncStatsChainHeight = height
	d.syncStatsLock.Unlock()
	d.committed.Store(true)

	// Initiate the sync using a concurrent header and content retrieval algorithm
	d.queue.Prepare(origin+1, mode)
	if d.syncInitHook != nil {
		d.syncInitHook(origin, height)
	}
	var headerFetcher func() error

	// headers are retrieved from the network
	headerFetcher = func() error { return d.fetchHeaders(p, origin+1, latest.Number.Uint64()) }

	fetchers := []func() error{
		headerFetcher, // Headers are always retrieved
		func() error { return d.fetchBodies(origin+1, beaconMode) },   // Bodies are retrieved during normal and snap sync
		func() error { return d.fetchReceipts(origin+1, beaconMode) }, // Receipts are retrieved during snap sync
		func() error { return d.processHeaders(origin+1, td, ttd, beaconMode) },
	}

	fetchers = append(fetchers, func() error { return d.processFullSyncContent(ttd, false) })
	return d.spawnSync(fetchers)
}

// spawnSync runs d.process and all given fetcher functions to completion in
// separate goroutines, returning the first error that appears.
func (d *Downloader) spawnSync(fetchers []func() error) error {
	errc := make(chan error, len(fetchers))
	d.cancelWg.Add(len(fetchers))
	for _, fn := range fetchers {
		fn := fn
		go func() { defer d.cancelWg.Done(); errc <- fn() }()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(fetchers); i++ {
		if i == len(fetchers)-1 {
			// Close the queue when all fetchers have exited.
			// This will cause the block processor to end when
			// it has processed the queue.
			d.queue.Close()
		}
		if got := <-errc; got != nil {
			err = got
			if got != errCanceled {
				break // receive a meaningful error, bubble it up
			}
		}
	}
	d.queue.Close()
	d.Cancel()
	return err
}

// cancel aborts all operations and resets the queue. However, cancel does
// not wait for the running download goroutines to finish. This method should be
// used when cancelling the downloads from inside the downloader.
func (d *Downloader) cancel() {
	// Close the current cancel channel
	if !d.cancelLock.TryLock() {
		panic("Failed to acquire cancelLock (downloader.go) in cancel()")
	}
	defer d.cancelLock.Unlock()

	if d.cancelCh != nil {
		select {
		case <-d.cancelCh:
			// Channel was already closed
		default:
			close(d.cancelCh)
		}
	}
}

// Cancel aborts all operations and waits for all download goroutines to
// finish before returning.
func (d *Downloader) Cancel() {
	d.cancel()
	d.cancelWg.Wait()
}

// Terminate interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (d *Downloader) Terminate() {
	// Close the termination channel (make sure double close is allowed)
	d.quitLock.Lock()
	select {
	case <-d.quitCh:
	default:
		close(d.quitCh)

		// Terminate the internal beacon syncer
		d.skeleton.Terminate()
	}
	d.quitLock.Unlock()

	// Cancel any pending download requests
	d.Cancel()
}

// fetchHead retrieves the head header and prior pivot block (if available) from
// a remote peer.
func (d *Downloader) fetchHead(p *peerConnection) (head *types.Header, pivot *types.Header, err error) {
	p.log.Debug("Retrieving remote chain head")

	// Request the advertised remote head block and wait for the response
	latest, _ := p.peer.Head()
	fetch := 1
	headers, hashes, err := d.fetchHeadersByHash(p, latest, fetch, fsMinFullBlocks-1, true)
	if err != nil {
		return nil, nil, err
	}
	// Make sure the peer gave us at least one and at most the requested headers
	if len(headers) == 0 || len(headers) > fetch {
		return nil, nil, fmt.Errorf("%w: returned headers %d != requested %d", errBadPeer, len(headers), fetch)
	}
	// The first header needs to be the head, validate against the request. If
	// only 1 header was returned, make sure there's no pivot or there was not
	// one requested.
	head = headers[0]
	if len(headers) == 1 {
		p.log.Debug("Remote head identified, no pivot", "number", head.Number, "hash", hashes[0])
		return head, nil, nil
	}
	// At this point we have 2 headers in total and the first is the
	// validated head of the chain. Check the pivot number and return,
	pivot = headers[1]
	if pivot.Number.Uint64() != head.Number.Uint64()-uint64(fsMinFullBlocks) {
		return nil, nil, fmt.Errorf("%w: remote pivot %d != requested %d", errInvalidChain, pivot.Number, head.Number.Uint64()-uint64(fsMinFullBlocks))
	}
	return head, pivot, nil
}

// calculateRequestSpan calculates what headers to request from a peer when trying to determine the
// common ancestor.
// It returns parameters to be used for peer.RequestHeadersByNumber:
//
//	from  - starting block number
//	count - number of headers to request
//	skip  - number of headers to skip
//
// and also returns 'max', the last block which is expected to be returned by the remote peers,
// given the (from,count,skip)
func calculateRequestSpan(remoteHeight, localHeight uint64) (int64, int, int, uint64) {
	var (
		from     int
		count    int
		MaxCount = MaxHeaderFetch / 16
	)
	// requestHead is the highest block that we will ask for. If requestHead is not offset,
	// the highest block that we will get is 16 blocks back from head, which means we
	// will fetch 14 or 15 blocks unnecessarily in the case the height difference
	// between us and the peer is 1-2 blocks, which is most common
	requestHead := int(remoteHeight) - 1
	if requestHead < 0 {
		requestHead = 0
	}
	// requestBottom is the lowest block we want included in the query
	// Ideally, we want to include the one just below our own head
	requestBottom := int(localHeight - 1)
	if requestBottom < 0 {
		requestBottom = 0
	}
	totalSpan := requestHead - requestBottom
	span := 1 + totalSpan/MaxCount
	if span < 2 {
		span = 2
	}
	if span > 16 {
		span = 16
	}

	count = 1 + totalSpan/span
	if count > MaxCount {
		count = MaxCount
	}
	if count < 2 {
		count = 2
	}
	from = requestHead - (count-1)*span
	if from < 0 {
		from = 0
	}
	max := from + (count-1)*span
	return int64(from), count, span - 1, uint64(max)
}

// findAncestor tries to locate the common ancestor link of the local chain and
// a remote peers blockchain. In the general case when our node was in sync and
// on the correct chain, checking the top N links should already get us a match.
// In the rare scenario when we ended up on a long reorganisation (i.e. none of
// the head links match), we do a binary search to find the common ancestor.
func (d *Downloader) findAncestor(p *peerConnection, remoteHeader *types.Header) (uint64, error) {
	// Figure out the valid ancestor range to prevent rewrite attacks
	var (
		floor        = int64(-1)
		localHeight  uint64
		remoteHeight = remoteHeader.Number.Uint64()
	)
	localHeight = d.blockchain.CurrentBlock().Number.Uint64()
	p.log.Debug("Looking for common ancestor", "local", localHeight, "remote", remoteHeight)

	// Recap floor value for binary search
	maxForkAncestry := fullMaxForkAncestry

	if localHeight >= maxForkAncestry {
		// We're above the max reorg threshold, find the earliest fork point
		floor = int64(localHeight - maxForkAncestry)
	}

	ancestor, err := d.findAncestorSpanSearch(p, FullSync, remoteHeight, localHeight, floor)
	if err == nil {
		return ancestor, nil
	}
	// The returned error was not nil.
	// If the error returned does not reflect that a common ancestor was not found, return it.
	// If the error reflects that a common ancestor was not found, continue to binary search,
	// where the error value will be reassigned.
	if !errors.Is(err, errNoAncestorFound) {
		return 0, err
	}

	ancestor, err = d.findAncestorBinarySearch(p, FullSync, remoteHeight, floor)
	if err != nil {
		return 0, err
	}
	return ancestor, nil
}

func (d *Downloader) findAncestorSpanSearch(p *peerConnection, mode SyncMode, remoteHeight, localHeight uint64, floor int64) (uint64, error) {
	from, count, skip, max := calculateRequestSpan(remoteHeight, localHeight)

	p.log.Trace("Span searching for common ancestor", "count", count, "from", from, "skip", skip)
	headers, hashes, err := d.fetchHeadersByNumber(p, uint64(from), count, skip, false)
	if err != nil {
		return 0, err
	}
	// Wait for the remote response to the head fetch
	number, hash := uint64(0), common.Hash{}

	// Make sure the peer actually gave something valid
	if len(headers) == 0 {
		p.log.Warn("Empty head header set")
		return 0, errEmptyHeaderSet
	}
	// Make sure the peer's reply conforms to the request
	for i, header := range headers {
		expectNumber := from + int64(i)*int64(skip+1)
		if number := header.Number.Int64(); number != expectNumber {
			p.log.Warn("Head headers broke chain ordering", "index", i, "requested", expectNumber, "received", number)
			return 0, fmt.Errorf("%w: %v", errInvalidChain, errors.New("head headers broke chain ordering"))
		}
	}
	// Check if a common ancestor was found
	for i := len(headers) - 1; i >= 0; i-- {
		// Skip any headers that underflow/overflow our requested set
		if headers[i].Number.Int64() < from || headers[i].Number.Uint64() > max {
			continue
		}
		// Otherwise check if we already know the header or not
		h := hashes[i]
		n := headers[i].Number.Uint64()

		known := d.blockchain.HasBlock(h, n)
		if known {
			number, hash = n, h
			break
		}
	}
	// If the head fetch already found an ancestor, return
	if hash != (common.Hash{}) {
		if int64(number) <= floor {
			p.log.Warn("Ancestor below allowance", "number", number, "hash", hash, "allowance", floor)
			return 0, errInvalidAncestor
		}
		p.log.Debug("Found common ancestor", "number", number, "hash", hash)
		return number, nil
	}
	return 0, errNoAncestorFound
}

func (d *Downloader) findAncestorBinarySearch(p *peerConnection, mode SyncMode, remoteHeight uint64, floor int64) (uint64, error) {
	hash := common.Hash{}

	// Ancestor not found, we need to binary search over our chain
	start, end := uint64(0), remoteHeight
	if floor > 0 {
		start = uint64(floor)
	}
	p.log.Trace("Binary searching for common ancestor", "start", start, "end", end)

	for start+1 < end {
		// Split our chain interval in two, and request the hash to cross check
		check := (start + end) / 2

		headers, hashes, err := d.fetchHeadersByNumber(p, check, 1, 0, false)
		if err != nil {
			return 0, err
		}
		// Make sure the peer actually gave something valid
		if len(headers) != 1 {
			p.log.Warn("Multiple headers for single request", "headers", len(headers))
			return 0, fmt.Errorf("%w: multiple headers (%d) for single request", errBadPeer, len(headers))
		}
		// Modify the search interval based on the response
		h := hashes[0]
		n := headers[0].Number.Uint64()
		known := d.blockchain.HasBlock(h, n)

		if !known {
			end = check
			continue
		}
		header := d.lightchain.GetHeaderByHash(h) // Independent of sync mode, header surely exists
		if header.Number.Uint64() != check {
			p.log.Warn("Received non requested header", "number", header.Number, "hash", header.Hash(), "request", check)
			return 0, fmt.Errorf("%w: non-requested header (%d)", errBadPeer, header.Number)
		}
		start = check
		hash = h
	}
	// Ensure valid ancestry and return
	if int64(start) <= floor {
		p.log.Warn("Ancestor below allowance", "number", start, "hash", hash, "allowance", floor)
		return 0, errInvalidAncestor
	}
	p.log.Debug("Found common ancestor", "number", start, "hash", hash)
	return start, nil
}

// fetchHeaders keeps retrieving headers concurrently from the number
// requested, until no more are returned, potentially throttling on the way. To
// facilitate concurrency but still protect against malicious nodes sending bad
// headers, we construct a header chain skeleton using the "origin" peer we are
// syncing with, and fill in the missing headers using anyone else. Headers from
// other peers are only accepted if they map cleanly to the skeleton. If no one
// can fill in the skeleton - not even the origin peer - it's assumed invalid and
// the origin is dropped.
func (d *Downloader) fetchHeaders(p *peerConnection, from uint64, head uint64) error {
	p.log.Debug("Directing header downloads", "origin", from)
	defer p.log.Debug("Header download terminated")

	// Start pulling the header chain skeleton until all is done
	var (
		skeleton = true  // Skeleton assembly phase or finishing up
		pivoting = false // Whether the next request is pivot verification
		ancestor = from
	)
	for {
		// Pull the next batch of headers, it either:
		//   - Pivot check to see if the chain moved too far
		//   - Skeleton retrieval to permit concurrent header fetches
		//   - Full header retrieval if we're near the chain head
		var (
			headers []*types.Header
			hashes  []common.Hash
			err     error
		)
		switch {
		case pivoting:
			d.pivotLock.RLock()
			pivot := d.pivotHeader.Number.Uint64()
			d.pivotLock.RUnlock()

			p.log.Trace("Fetching next pivot header", "number", pivot+uint64(fsMinFullBlocks))
			headers, hashes, err = d.fetchHeadersByNumber(p, pivot+uint64(fsMinFullBlocks), 2, fsMinFullBlocks-9, false) // move +64 when it's 2x64-8 deep

		case skeleton:
			p.log.Trace("Fetching skeleton headers", "count", MaxHeaderFetch, "from", from)
			headers, hashes, err = d.fetchHeadersByNumber(p, from+uint64(MaxHeaderFetch)-1, MaxSkeletonSize, MaxHeaderFetch-1, false)

		default:
			p.log.Trace("Fetching full headers", "count", MaxHeaderFetch, "from", from)
			headers, hashes, err = d.fetchHeadersByNumber(p, from, MaxHeaderFetch, 0, false)
		}
		switch err {
		case nil:
			// Headers retrieved, continue with processing

		case errCanceled:
			// Sync cancelled, no issue, propagate up
			return err

		default:
			// Header retrieval either timed out, or the peer failed in some strange way
			// (e.g. disconnect). Consider the master peer bad and drop
			d.dropPeer(p.id)

			// Finish the sync gracefully instead of dumping the gathered data though
			for _, ch := range []chan bool{d.queue.blockWakeCh, d.queue.receiptWakeCh} {
				select {
				case ch <- false:
				case <-d.cancelCh:
				}
			}
			select {
			case d.headerProcCh <- nil:
			case <-d.cancelCh:
			}
			return fmt.Errorf("%w: header request failed: %v", errBadPeer, err)
		}
		// If the pivot is being checked, move if it became stale and run the real retrieval
		var pivot uint64

		d.pivotLock.RLock()
		if d.pivotHeader != nil {
			pivot = d.pivotHeader.Number.Uint64()
		}
		d.pivotLock.RUnlock()

		if pivoting {
			if len(headers) == 2 {
				if have, want := headers[0].Number.Uint64(), pivot+uint64(fsMinFullBlocks); have != want {
					log.Warn("Peer sent invalid next pivot", "have", have, "want", want)
					return fmt.Errorf("%w: next pivot number %d != requested %d", errInvalidChain, have, want)
				}
				if have, want := headers[1].Number.Uint64(), pivot+2*uint64(fsMinFullBlocks)-8; have != want {
					log.Warn("Peer sent invalid pivot confirmer", "have", have, "want", want)
					return fmt.Errorf("%w: next pivot confirmer number %d != requested %d", errInvalidChain, have, want)
				}
				log.Warn("Pivot seemingly stale, moving", "old", pivot, "new", headers[0].Number)
				pivot = headers[0].Number.Uint64()

				d.pivotLock.Lock()
				d.pivotHeader = headers[0]
				d.pivotLock.Unlock()

				// Write out the pivot into the database so a rollback beyond
				// it will reenable snap sync and update the state root that
				// the state syncer will be downloading.
				rawdb.WriteLastPivotNumber(d.stateDB, pivot)
			}
			// Disable the pivot check and fetch the next batch of headers
			pivoting = false
			continue
		}
		// If the skeleton's finished, pull any remaining head headers directly from the origin
		if skeleton && len(headers) == 0 {
			// A malicious node might withhold advertised headers indefinitely
			if from+uint64(MaxHeaderFetch)-1 <= head {
				p.log.Warn("Peer withheld skeleton headers", "advertised", head, "withheld", from+uint64(MaxHeaderFetch)-1)
				return fmt.Errorf("%w: withheld skeleton headers: advertised %d, withheld #%d", errStallingPeer, head, from+uint64(MaxHeaderFetch)-1)
			}
			p.log.Debug("No skeleton, fetching headers directly")
			skeleton = false
			continue
		}
		// If no more headers are inbound, notify the content fetchers and return
		if len(headers) == 0 {
			// Don't abort header fetches while the pivot is downloading
			if !d.committed.Load() && pivot <= from {
				p.log.Debug("No headers, waiting for pivot commit")
				select {
				case <-time.After(fsHeaderContCheck):
					continue
				case <-d.cancelCh:
					return errCanceled
				}
			}
			// Pivot done (or not in snap sync) and no more headers, terminate the process
			p.log.Debug("No more headers available")
			select {
			case d.headerProcCh <- nil:
				return nil
			case <-d.cancelCh:
				return errCanceled
			}
		}
		// If we received a skeleton batch, resolve internals concurrently
		var progressed bool
		if skeleton {
			filled, hashset, proced, err := d.fillHeaderSkeleton(from, headers)
			if err != nil {
				p.log.Debug("Skeleton chain invalid", "err", err)
				return fmt.Errorf("%w: %v", errInvalidChain, err)
			}
			headers = filled[proced:]
			hashes = hashset[proced:]

			progressed = proced > 0
			from += uint64(proced)
		} else {
			// A malicious node might withhold advertised headers indefinitely
			if n := len(headers); n < MaxHeaderFetch && headers[n-1].Number.Uint64() < head {
				p.log.Warn("Peer withheld headers", "advertised", head, "delivered", headers[n-1].Number.Uint64())
				return fmt.Errorf("%w: withheld headers: advertised %d, delivered %d", errStallingPeer, head, headers[n-1].Number.Uint64())
			}
			// If we're closing in on the chain head, but haven't yet reached it, delay
			// the last few headers so mini reorgs on the head don't cause invalid hash
			// chain errors.
			if n := len(headers); n > 0 {
				// Retrieve the current head we're at
				head := d.blockchain.CurrentSnapBlock().Number.Uint64()
				if full := d.blockchain.CurrentBlock().Number.Uint64(); head < full {
					head = full
				}

				// If the head is below the common ancestor, we're actually deduplicating
				// already existing chain segments, so use the ancestor as the fake head.
				// Otherwise, we might end up delaying header deliveries pointlessly.
				if head < ancestor {
					head = ancestor
				}
				// If the head is way older than this batch, delay the last few headers
				if head+uint64(reorgProtThreshold) < headers[n-1].Number.Uint64() {
					delay := reorgProtHeaderDelay
					if delay > n {
						delay = n
					}
					headers = headers[:n-delay]
					hashes = hashes[:n-delay]
				}
			}
		}
		// If no headers have been delivered, or all of them have been delayed,
		// sleep a bit and retry. Take care with headers already consumed during
		// skeleton filling
		if len(headers) == 0 && !progressed {
			p.log.Trace("All headers delayed, waiting")
			select {
			case <-time.After(fsHeaderContCheck):
				continue
			case <-d.cancelCh:
				return errCanceled
			}
		}
		// Insert any remaining new headers and fetch the next batch
		if len(headers) > 0 {
			p.log.Trace("Scheduling new headers", "count", len(headers), "from", from)
			select {
			case d.headerProcCh <- &headerTask{
				headers: headers,
				hashes:  hashes,
			}:
			case <-d.cancelCh:
				return errCanceled
			}
			from += uint64(len(headers))
		}
		// If we're still skeleton filling snap sync, check pivot staleness
		// before continuing to the next skeleton filling
		if skeleton && pivot > 0 {
			pivoting = true
		}
	}
}

// fillHeaderSkeleton concurrently retrieves headers from all our available peers
// and maps them to the provided skeleton header chain.
//
// Any partial results from the beginning of the skeleton is (if possible) forwarded
// immediately to the header processor to keep the rest of the pipeline full even
// in the case of header stalls.
//
// The method returns the entire filled skeleton and also the number of headers
// already forwarded for processing.
func (d *Downloader) fillHeaderSkeleton(from uint64, skeleton []*types.Header) ([]*types.Header, []common.Hash, int, error) {
	log.Debug("Filling up skeleton", "from", from)
	d.queue.ScheduleSkeleton(from, skeleton)

	err := d.concurrentFetch((*headerQueue)(d), false)
	if err != nil {
		log.Debug("Skeleton fill failed", "err", err)
	}
	filled, hashes, proced := d.queue.RetrieveHeaders()
	if err == nil {
		log.Debug("Skeleton fill succeeded", "filled", len(filled), "processed", proced)
	}
	return filled, hashes, proced, err
}

// fetchBodies iteratively downloads the scheduled block bodies, taking any
// available peers, reserving a chunk of blocks for each, waiting for delivery
// and also periodically checking for timeouts.
func (d *Downloader) fetchBodies(from uint64, beaconMode bool) error {
	log.Debug("Downloading block bodies", "origin", from)
	err := d.concurrentFetch((*bodyQueue)(d), beaconMode)

	log.Debug("Block body download terminated", "err", err)
	return err
}

// fetchReceipts iteratively downloads the scheduled block receipts, taking any
// available peers, reserving a chunk of receipts for each, waiting for delivery
// and also periodically checking for timeouts.
func (d *Downloader) fetchReceipts(from uint64, beaconMode bool) error {
	log.Debug("Downloading receipts", "origin", from)
	err := d.concurrentFetch((*receiptQueue)(d), beaconMode)

	log.Debug("Receipt download terminated", "err", err)
	return err
}

// processHeaders takes batches of retrieved headers from an input channel and
// keeps processing and scheduling them into the header chain and downloader's
// queue until the stream ends or a failure occurs.
func (d *Downloader) processHeaders(origin uint64, td, ttd *big.Int, beaconMode bool) error {
	var (
		gotHeaders = false // Wait for batches of headers to process
	)
	for {
		select {
		case <-d.cancelCh:
			return errCanceled

		case task := <-d.headerProcCh:
			// Terminate header processing if we synced up
			if task == nil || len(task.headers) == 0 {
				// Notify everyone that headers are fully processed
				for _, ch := range []chan bool{d.queue.blockWakeCh, d.queue.receiptWakeCh} {
					select {
					case ch <- false:
					case <-d.cancelCh:
					}
				}
				// We need to check total difficulty violations from malicious peers.
				// If no headers were retrieved at all, the peer violated its TD promise that it had a
				// better chain compared to ours. The only exception is if its promised blocks were
				// already imported by other means (e.g. fetcher):
				//
				// R <remote peer>, L <local node>: Both at block 10
				// R: Mine block 11, and propagate it to L
				// L: Queue block 11 for import
				// L: Notice that R's head and TD increased compared to ours, start sync
				// L: Import of block 11 finishes
				// L: Sync begins, and finds common ancestor at 11
				// L: Request new headers up from 11 (R's TD was higher, it must have something)
				// R: Nothing to give
				head := d.blockchain.CurrentBlock()
				if !gotHeaders && td.Cmp(d.blockchain.GetTd(head.Hash(), head.Number.Uint64())) > 0 {
					return errStallingPeer
				}
				return nil
			}
			// Otherwise split the chunk of headers into batches and process them
			headers, hashes := task.headers, task.hashes

			gotHeaders = true
			for len(headers) > 0 {
				// Terminate if something failed in between processing chunks
				select {
				case <-d.cancelCh:
					return errCanceled
				default:
				}
				// Select the next chunk of headers to import
				limit := maxHeadersProcess
				if limit > len(headers) {
					limit = len(headers)
				}
				chunkHeaders := headers[:limit]
				chunkHashes := hashes[:limit]

				// If we've reached the allowed number of pending headers, stall a bit
				for d.queue.PendingBodies() >= maxQueuedHeaders || d.queue.PendingReceipts() >= maxQueuedHeaders {
					select {
					case <-d.cancelCh:
						return errCanceled
					case <-time.After(100 * time.Millisecond):
					}
				}
				// Otherwise insert the headers for content retrieval
				inserts := d.queue.Schedule(chunkHeaders, chunkHashes, origin)
				if len(inserts) != len(chunkHeaders) {
					return fmt.Errorf("%w: stale headers", errBadPeer)
				}

				headers = headers[limit:]
				hashes = hashes[limit:]
				origin += uint64(limit)
			}
			// Update the highest block number we know if a higher one is found.
			d.syncStatsLock.Lock()
			if d.syncStatsChainHeight < origin {
				d.syncStatsChainHeight = origin - 1
			}
			d.syncStatsLock.Unlock()

			// Signal the content downloaders on the availability of new tasks
			for _, ch := range []chan bool{d.queue.blockWakeCh, d.queue.receiptWakeCh} {
				select {
				case ch <- true:
				default:
				}
			}
		}
	}
}

// processFullSyncContent takes fetch results from the queue and imports them into the chain.
func (d *Downloader) processFullSyncContent(ttd *big.Int, beaconMode bool) error {
	for {
		results := d.queue.Results(true)
		if len(results) == 0 {
			return nil
		}
		if d.chainInsertHook != nil {
			d.chainInsertHook(results)
		}
		// Although the received blocks might be all valid, a legacy PoW/PoA sync
		// must not accept post-merge blocks. Make sure that pre-merge blocks are
		// imported, but post-merge ones are rejected.
		var (
			rejected []*fetchResult
			td       *big.Int
		)

		td = d.blockchain.GetTd(results[0].Header.ParentHash, results[0].Header.Number.Uint64()-1)
		if td == nil {
			// This should never really happen, but handle gracefully for now
			log.Error("Failed to retrieve parent block TD", "number", results[0].Header.Number.Uint64()-1, "hash", results[0].Header.ParentHash)
			return fmt.Errorf("%w: parent TD missing", errInvalidChain)
		}

		if ttd != nil {
			for i, result := range results {
				td = new(big.Int).Add(td, result.Header.Difficulty)
				if td.Cmp(ttd) >= 0 {
					// Terminal total difficulty reached, allow the last block in
					if new(big.Int).Sub(td, result.Header.Difficulty).Cmp(ttd) < 0 {
						results, rejected = results[:i+1], results[i+1:]
						if len(rejected) > 0 {
							// Make a nicer user log as to the first TD truly rejected
							td = new(big.Int).Add(td, rejected[0].Header.Difficulty)
						}
					} else {
						results, rejected = results[:i], results[i:]
					}
					break
				}
			}
		}

		if err := d.importBlockResults(results); err != nil {
			return err
		}
	}
}

func (d *Downloader) importBlockResults(results []*fetchResult) error {
	// Check for any early termination requests
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	default:
	}
	// Retrieve a batch of results to import
	first, last := results[0].Header, results[len(results)-1].Header
	log.Debug("Inserting downloaded chain", "items", len(results),
		"firstnum", first.Number, "firsthash", first.Hash(),
		"lastnum", last.Number, "lasthash", last.Hash(),
	)
	blocks := make([]*types.Block, len(results))
	for i, result := range results {
		blocks[i] = types.NewBlockWithHeader(result.Header).WithBody(result.Transactions, result.Uncles).WithWithdrawals(result.Withdrawals)
	}
	// Downloaded blocks are always regarded as trusted after the
	// transition. Because the downloaded chain is guided by the
	// consensus-layer.
	if index, err := d.blockchain.InsertChain(blocks); err != nil {
		if index < len(results) {
			log.Debug("Downloaded item processing failed", "number", results[index].Header.Number, "hash", results[index].Header.Hash(), "err", err)
		} else {
			// The InsertChain method in blockchain.go will sometimes return an out-of-bounds index,
			// when it needs to preprocess blocks to import a sidechain.
			// The importer will put together a new list of blocks to import, which is a superset
			// of the blocks delivered from the downloader, and the indexing will be off.
			log.Debug("Downloaded item processing failed on sidechain import", "index", index, "err", err)
		}
		return fmt.Errorf("%w: %v", errInvalidChain, err)
	}
	return nil
}

// readHeaderRange returns a list of headers, using the given last header as the base,
// and going backwards towards genesis. This method assumes that the caller already has
// placed a reasonable cap on count.
func (d *Downloader) readHeaderRange(last *types.Header, count int) []*types.Header {
	var (
		current = last
		headers []*types.Header
	)
	for {
		parent := d.lightchain.GetHeaderByHash(current.ParentHash)
		if parent == nil {
			break // The chain is not continuous, or the chain is exhausted
		}
		headers = append(headers, parent)
		if len(headers) >= count {
			break
		}
		current = parent
	}
	return headers
}
