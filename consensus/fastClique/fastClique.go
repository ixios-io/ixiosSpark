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

package fastClique

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ixios-io/ixiosSpark/accounts"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
	lru "github.com/ixios-io/ixiosSpark/common/lru"
	"github.com/ixios-io/ixiosSpark/consensus"
	"github.com/ixios-io/ixiosSpark/consensus/misc"
	"github.com/ixios-io/ixiosSpark/core/state"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/crypto"
	"github.com/ixios-io/ixiosSpark/kvdb"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/ixios-io/ixiosSpark/rlp"
	"github.com/ixios-io/ixiosSpark/rpc"
	"github.com/ixios-io/ixiosSpark/trie"
	"golang.org/x/crypto/sha3"
)

// FastClique Protocol constants.
const (
	checkpointInterval = 2048                   // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128                    // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096                   // Number of recent block signatures to keep in memory
	extraMKS           = 2656                   // Reserve 2690 bytes for Multi-Key-Signature (MKS)
	extraVanity        = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal          = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal
	maxBlocksOOT       = 3                      // Maximum number of blocks a validator can sign out-of-turn
	ootWaitMinimum     = 3500
	ootWaitLowerBound  = 6000
	ootWaitUpperBound  = 8500
	ootWaitMaximum     = 16500
)

var (
	epochLength = uint64(10000) // Default number of blocks after which to checkpoint and reset the pending votes

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcOmmerHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	diffInTurn = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 32 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidOmmerHash is returned if a block contains an non-empty uncle list.
	errInvalidOmmerHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errWrongDifficulty is returned if the difficulty of a block doesn't match the
	// turn of the signer.
	errWrongDifficulty = errors.New("wrong difficulty")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned = errors.New("recently signed")

	errInvalidMKSSize = errors.New("extra-data has incorrect MKS region size")
)

// SignerFn hashes and signs the data to be signed by a backing account.
type SignerFn func(signer accounts.Account, mimeType string, message []byte) ([]byte, error)

// ecrecover extracts the Ixios account address from a signed header.
func ecrecover(header *types.Header, sigcache *sigLRU) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address, nil
	}

	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ixios address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}

	var signer common.Address
	fullHash := crypto.Keccak256(pubkey[1:])
	copy(signer[:], fullHash)

	// Zero out first 6 bytes of the signer address
	for i := 0; i < 6; i++ {
		signer[i] = 0
	}

	log.Trace("Clique ecrecover details",
		"block", header.Number,
		"recovered_signer", signer,
		"header_hash", header.Hash(),
		"seal_hash", SealHash(header),
		"sig_len", len(signature))

	sigcache.Add(hash, signer)
	return signer, nil
}

type Clique struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     kvdb.Database        // Database to store and retrieve snapshot checkpoints

	recents    *lru.Cache[common.Hash, *Snapshot] // Snapshots for recent block to speed up reorgs
	signatures *sigLRU                            // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	signer common.Address // Ixios address of the signing key
	signFn SignerFn       // Signer function to authorize hashes with
	lock   sync.RWMutex   // Protects the signer and proposals fields

	// The fields below are for testing only
	fakeDiff bool // Skip difficulty verifications
}

// New creates a Clique proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.CliqueConfig, db kvdb.Database) *Clique {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents := lru.NewCache[common.Hash, *Snapshot](inmemorySnapshots)
	signatures := lru.NewCache[common.Hash, common.Address](inmemorySignatures)

	return &Clique{
		config:     &conf,
		db:         db,
		recents:    recents,
		signatures: signatures,
		proposals:  make(map[common.Address]bool),
	}
}

// Author implements consensus.Engine, returning the Ixios address recovered
// from the signature in the header's extra-data section.
func (c *Clique) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Clique) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Clique) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *Clique) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks more than 500ms into the future
	if header.Time > uint64(time.Now().UnixMilli())+500 {
		return consensus.ErrFutureBlock
	}
	checkpoint := (number % c.config.Epoch) == 0
	// Checkpoint blocks do not need to enforce zero beneficiary
	/*
		if checkpoint && header.Coinbase != (common.Address{}) {
			return errInvalidCheckpointBeneficiary
		}*/
	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}
	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}
	if checkpoint && signersBytes%common.AddressLength != 0 {
		return errInvalidCheckpointSigners
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.OmmerHash != uncleHash {
		return errInvalidOmmerHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return errInvalidDifficulty
		}
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}
	if chain.Config().IsShanghai(header.Number, header.Time) {
		return errors.New("clique does not support shanghai fork")
	}
	// Verify the non-existence of withdrawalsHash.
	if header.WithdrawalsHash != nil {
		return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
	}
	if chain.Config().IsCancun(header.Number, header.Time) {
		return errors.New("clique does not support cancun fork")
	}
	// Verify the non-existence of cancun-specific header fields
	switch {
	case header.ExcessBlobGas != nil:
		return fmt.Errorf("invalid excessBlobGas: have %d, expected nil", header.ExcessBlobGas)
	case header.BlobGasUsed != nil:
		return fmt.Errorf("invalid blobGasUsed: have %d, expected nil", header.BlobGasUsed)
	case header.ParentBeaconRoot != nil:
		return fmt.Errorf("invalid parentBeaconRoot, have %#x, expected nil", header.ParentBeaconRoot)
	}

	// Determine how many bytes are left for signers & MKS
	// The last extraSeal bytes are always the MKS portion.
	// We'll also subtract our known extraMKS length.
	//extraSuffix := len(header.Extra) - extraSeal
	//mksStart := extraSuffix - extraMKS
	// todo: implement MKS here
	// todo: errInvalidMKSSize

	// All basic checks passed, verify cascading fields
	return c.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *Clique) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}
	// If the block is a checkpoint block, verify the signer list

	// All basic checks passed, verify the seal and return
	return c.verifySeal(snap, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *Clique) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%checkpointInterval == 0 {
			if s, err := loadSnapshot(c.config, c.signatures, c.db, hash); err == nil {
				log.Trace("Loaded voting snapshot from disk", "number", number, "hash", hash)
				snap = s
				break
			}
		}
		// If we're at the genesis, snapshot the initial state. Alternatively if we're
		// at a checkpoint block without a parent (light client CHT), or we have piled
		// up more headers than allowed to be reorged (chain reinit from a freezer),
		// consider the checkpoint trusted and snapshot it.
		if number == 0 || (number%c.config.Epoch == 0 && (len(headers) > params.FullImmutabilityThreshold || chain.GetHeaderByNumber(number-1) == nil)) {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()

				// MKS Implementation
				// Extract signers from the genesis or checkpoint block's extra data
				// Each signer entry consists of the address (32 bytes) followed by MKS data (2690 bytes)
				signerEntrySize := common.AddressLength + extraMKS
				signers := make([]common.Address, (len(checkpoint.Extra)-extraVanity-extraSeal)/signerEntrySize)
				for i := 0; i < len(signers); i++ {
					// Extract only the address part, skipping the MKS data for each signer
					startPos := extraVanity + (i * signerEntrySize)
					copy(signers[i][:], checkpoint.Extra[startPos:startPos+common.AddressLength])
				}
				snap = newSnapshot(c.config, c.signatures, number, hash, signers)
				if err := snap.store(c.db); err != nil {
					return nil, err
				}

				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}
	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = snap.store(c.db); err != nil {
			return nil, err
		}
		log.Trace("Stored voting snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Clique) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *Clique) verifySeal(snap *Snapshot, header *types.Header, parents []*types.Header) error {
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}

	signer, err := ecrecover(header, c.signatures)

	if err != nil {
		log.Debug("Clique ecrecover failed",
			"block", number,
			"error", err)
		return err
	}

	// Check authorized signers
	authorized := false
	for authSigner := range snap.Signers {
		// Check for zero prefix in authorized signer
		hasZeroPrefix := true
		zeroPrefix := make([]byte, 12) // Creates 12 zero bytes
		if !bytes.Equal(authSigner[:12], zeroPrefix) {
			hasZeroPrefix = false
		}

		log.Trace("Clique checking signer",
			"block", number,
			"auth_signer", authSigner,
			"has_zero_prefix", hasZeroPrefix,
			"last_20_equal", bytes.Equal(signer[12:], authSigner[12:]),
			"full_equal", signer == authSigner)

		// For zero-prefixed authorized signers, compare only the last 20 bytes
		if hasZeroPrefix && bytes.Equal(signer[12:], authSigner[12:]) {
			log.Trace("Clique authorized via last 20 bytes",
				"block", number,
				"signer", signer,
				"auth_signer", authSigner)
			authorized = true
			break
		}

		// For full 32-byte authorized signers, compare everything
		if !hasZeroPrefix && signer == authSigner {
			log.Debug("Clique authorized via full match",
				"block", number,
				"signer", signer)
			authorized = true
			break
		}
	}

	if !authorized {
		return errUnauthorizedSigner
	}

	// Count recent blocks by each validator
	recentBlocks := countRecentBlocksByValidator(snap, number)

	// Check if signer has exceeded their block limit
	inturn := snap.inturn(number, signer)
	if !inturn {
		if len(recentBlocks[signer]) >= maxBlocksOOT {
			return errRecentlySigned
		}
	}

	// Check difficulty
	if !c.fakeDiff {
		if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
			return errWrongDifficulty
		}
		if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
			return errWrongDifficulty
		}
	}

	return nil
}

// getOutOfTurnootWait returns a random ootWait time between 1500ms and 9500ms for out-of-turn signers.
func getOutOfTurnootWait() time.Duration {
	if rand.Int()%3 == 0 {
		return time.Duration(ootWaitMinimum+rand.Intn(ootWaitLowerBound)) * time.Millisecond
	} else {
		return time.Duration((ootWaitMaximum-ootWaitUpperBound)+rand.Intn(ootWaitUpperBound)) * time.Millisecond
	}
}

// hasBlockArrived checks if a block at or above the given targetNumber now exists.
func hasBlockArrived(chain consensus.ChainHeaderReader, targetNumber uint64) bool {
	latestHeader := chain.CurrentHeader()
	if latestHeader == nil {
		return false
	}
	return latestHeader.Number.Uint64() >= targetNumber
}

// Helper function to get the out-of-turn block limit window size
func getOutOfTurnBlockLimit(numValidators int) uint64 {
	return uint64(numValidators * 2)
}

// Helper to count recent blocks signed by each validator
func countRecentBlocksByValidator(snap *Snapshot, currentBlock uint64) map[common.Address][]uint64 {
	recentBlocks := make(map[common.Address][]uint64)

	// Initialize slice for each signer
	for signer := range snap.Signers {
		recentBlocks[signer] = make([]uint64, 0)
	}

	// Get window size based on number of validators
	windowSize := getOutOfTurnBlockLimit(len(snap.Signers))

	// Collect block numbers signed by each validator within window
	for blockNum, signer := range snap.Recents {
		if currentBlock-blockNum <= windowSize {
			recentBlocks[signer] = append(recentBlocks[signer], blockNum)
		}
	}

	return recentBlocks
}

// Seal tries to create a sealed block using the local signing credentials, incorporating
// randomized out-of-turn ootWait times and conditional waiting to reduce ommers and maintain
// near 1-second block times even with large validator sets.
func (c *Clique) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := block.Header()
	number := header.Number.Uint64()

	if number == 0 {
		return errUnknownBlock
	}

	c.lock.RLock()
	signer, signFn := c.signer, c.signFn
	c.lock.RUnlock()
	log.Debug("Begin Sealing block")

	// Get snapshot and do initial checks
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		log.Error("Failed to seal block: Failed to get snapshot", "error", err)
		return err
	}

	// Authorization check
	if _, authorized := snap.Signers[signer]; !authorized {
		log.Error("Failed to seal block: Signer not authorized",
			"signer", signer,
			"authorizedCount", len(snap.Signers))
		return errUnauthorizedSigner
	}

	// Get block signing history
	recentBlocks := countRecentBlocksByValidator(snap, number)

	// Determine if we're in turn
	inTurn := snap.inturn(number, signer)

	// Check if we can sign out of turn
	if !inTurn {
		myRecentBlocks := recentBlocks[signer]
		if len(myRecentBlocks) >= maxBlocksOOT {
			log.Warn("Cannot seal - reached maximum out-of-turn blocks",
				"recent_blocks", myRecentBlocks,
				"maxAllowed", maxBlocksOOT)

			return errors.New("max out-of-turn blocks reached")
		}
	}

	// ootWaitTime & Delay
	delay := 0 * time.Millisecond
	ootWaitTime := getOutOfTurnootWait()

	if inTurn {
		parent := chain.GetHeader(header.ParentHash, number-1)
		if parent == nil {
			return consensus.ErrUnknownAncestor
		}

		// Calculate how long it's been since parent was created
		parentAge := time.Now().UnixMilli() - int64(parent.Time)
		if parentAge < int64(c.config.Period) {
			delay = (time.Duration(c.config.Period-uint64(parentAge)) * time.Millisecond)
		}
		log.Debug("Calculated delay",
			"parentAge", parentAge,
			"targetBlockTime", c.config.Period,
			"delay", delay,
			"parentTime", parent.Time,
			"now", time.Now().UnixMilli())
	} else {
		parent := chain.GetHeader(header.ParentHash, number-1)
		if parent == nil {
			return consensus.ErrUnknownAncestor
		}
		parentAge := time.Now().UnixMilli() - int64(parent.Time)
		if uint64(parentAge) < c.config.Period {
			delay = time.Duration(c.config.Period-uint64(parentAge))*time.Millisecond + ootWaitTime
		} else {
			delay = ootWaitTime
		}
	}

	// Launch sealing goroutine
	go func() {
		// Calculate delays
		if !inTurn {
			log.Debug("Waiting before sealing out-of-turn",
				"block", number, "delay", delay,
				"in_turn", inTurn)
		} else {
			log.Info("It's our turn to seal the block", "block", number)
		}

		// Wait for the delay
		select {
		case <-time.After(delay):
		case <-stop:
			return
		}
		// Check if block arrived during wait
		if !inTurn && hasBlockArrived(chain, number) {
			log.Debug("Not sealing OOT, block arrived during wait.")
			return
		}

		// Double check recent blocks after delay
		newSnap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
		if err != nil {
			log.Error("Failed to get new snapshot after delay", "error", err)
			return
		}

		newRecentBlocks := countRecentBlocksByValidator(newSnap, number)
		if (len(newRecentBlocks[signer]) + 1) >= maxBlocksOOT {
			log.Debug("Not sealing OOT, max out of turn blocks reached.")
			return
		}

		// Sign the block
		sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeClique, CliqueRLP(header))
		if err != nil {
			log.Error("Failed to sign block: signFn failed", "error", err)
			return
		}
		copy(header.Extra[len(header.Extra)-extraSeal:], sighash)

		// Check if block arrived during wait (final check)
		if !inTurn && hasBlockArrived(chain, number) {
			log.Debug("Not sealing OOT, block arrived during wait.")
			return
		}

		// Send the sealed block
		select {
		case results <- block.WithSeal(header):
			//estReward := c.EstimateReward(chain, block)
			if !inTurn {
				log.Warn("Sealed out of turn, in-turn signer failed to sign", "block", number, "delay", delay)
			}
		case <-stop:
			return
		}
	}()

	return nil
}

func (c *Clique) EstimateReward(chain consensus.ChainHeaderReader, block *types.Block) *big.Int {
	totalGasFees := new(big.Int)
	txs := block.Transactions()
	for _, tx := range txs {
		// If the transaction has a gas price, approximate fees as (tx.Gas() * tx.GasPrice()).
		if tx.GasPrice() != nil {
			gasLimitBig := new(big.Int).SetUint64(tx.Gas())
			fees := new(big.Int).Mul(gasLimitBig, tx.GasPrice())
			totalGasFees.Add(totalGasFees, fees)
		}
	}

	return totalGasFees
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *Clique) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// Set coinbase to the local node's validator address (c.signer).
	header.Coinbase = c.signer

	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	c.lock.RLock()
	if number%c.config.Epoch != 0 {
		addresses := make([]common.Address, 0, len(c.proposals))
		for address, authorize := range c.proposals {
			if snap.validVote(address, authorize) {
				addresses = append(addresses, address)
			}
		}
		if len(addresses) > 0 {
			header.Coinbase = addresses[rand.Intn(len(addresses))]
			if c.proposals[header.Coinbase] {
				copy(header.Nonce[:], nonceAuthVote)
			} else {
				copy(header.Nonce[:], nonceDropVote)
			}
		}
	}
	signer := c.signer
	c.lock.RUnlock()

	header.Difficulty = calcDifficulty(snap, signer)

	// Ensure the extra-data has the right layout
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]
	if number%c.config.Epoch == 0 {
		for _, s := range snap.signers() {
			header.Extra = append(header.Extra, s[:]...)
		}
	}
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)

	// Mix digest is reserved/unused, set to empty
	header.MixDigest = common.Hash{}

	// Bump timestamp so we don't go backwards in real time
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	targetTime := parent.Time + c.config.Period
	now := uint64(time.Now().UnixMilli())
	if targetTime < now {
		header.Time = now
	} else {
		header.Time = targetTime
	}
	return nil
}

// Finalize implements consensus.Engine. There is no post-transaction
// consensus rules in clique, do nothing here.
func (c *Clique) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, withdrawals []*types.Withdrawal) {
	// No block rewards in PoA, so the state remains as is
}

func (c *Clique) FinalizeAndAssemble(
	chain consensus.ChainHeaderReader,
	header *types.Header,
	state *state.StateDB,
	txs []*types.Transaction,
	ommers []*types.Header,
	receipts []*types.Receipt,
	withdrawals []*types.Withdrawal,
) (*types.Block, error) {
	if len(withdrawals) > 0 {
		return nil, errors.New("withdrawals are not supported")
	}
	// Finalise the block
	c.Finalize(chain, header, state, txs, ommers, withdrawals)

	// Assign the final state root to the header.
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Return the final block for sealing.
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *Clique) Authorize(signer common.Address, signFn SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.signer = signer
	c.signFn = signFn
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have:
// * DIFF_NOTURN(2) if BLOCK_NUMBER % SIGNER_COUNT != SIGNER_INDEX
// * DIFF_INTURN(1) if BLOCK_NUMBER % SIGNER_COUNT == SIGNER_INDEX
func (c *Clique) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	snap, err := c.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
	if err != nil {
		return nil
	}
	c.lock.RLock()
	signer := c.signer
	c.lock.RUnlock()
	return calcDifficulty(snap, signer)
}

func calcDifficulty(snap *Snapshot, signer common.Address) *big.Int {
	if snap.inturn(snap.Number+1, signer) {
		return new(big.Int).Set(diffInTurn)
	}
	return new(big.Int).Set(diffNoTurn)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Clique) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *Clique) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *Clique) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "clique",
		Service:   &API{chain: chain, clique: c},
	}}
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

// CliqueRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func CliqueRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.OmmerHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if header.WithdrawalsHash != nil {
		panic("unexpected withdrawal hash value in clique")
	}
	if header.ExcessBlobGas != nil {
		panic("unexpected excess blob gas value in clique")
	}
	if header.BlobGasUsed != nil {
		panic("unexpected blob gas used value in clique")
	}
	if header.ParentBeaconRoot != nil {
		panic("unexpected parent beacon root value in clique")
	}
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}
