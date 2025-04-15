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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/rawdb"
	"github.com/ixios-io/ixiosSpark/core/state"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/crypto"
	"github.com/ixios-io/ixiosSpark/internal/era"
	"github.com/ixios-io/ixiosSpark/internal/flags"
	"github.com/ixios-io/ixiosSpark/kvdb"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/node"
	"github.com/ixios-io/ixiosSpark/params"
	"github.com/urfave/cli/v2"
)

var (
	dumpGenesisCommand = &cli.Command{
		Action:    dumpGenesis,
		Name:      "dumpgenesis",
		Usage:     "Dumps genesis block JSON configuration to stdout",
		ArgsUsage: "",
		Flags:     append([]cli.Flag{DataDirFlag}, NetworkFlags...),
		Description: `
The dumpgenesis command prints the genesis configuration of the network preset
if one is set.  Otherwise it prints the genesis from the datadir.`,
	}
	importCommand = &cli.Command{
		Action:    importChain,
		Name:      "import",
		Usage:     "Import a blockchain file",
		ArgsUsage: "<filename> (<filename 2> ... <filename N>) ",
		Flags: flags.Merge([]cli.Flag{
			CacheFlag,
			SyncModeFlag,
			GCModeFlag,
			SnapshotFlag,
			CacheDatabaseFlag,
			CacheGCFlag,
			TransactionHistoryFlag,
			StateHistoryFlag,
		}, DatabaseFlags),
		Description: `
The import command imports blocks from an RLP-encoded form. The form can be one file
with several RLP-encoded blocks, or several files can be used.

If only one file is used, import error will result in failure. If several files are used,
processing will proceed even if an individual RLP-file import failure occurs.`,
	}
	exportCommand = &cli.Command{
		Action:    exportChain,
		Name:      "export",
		Usage:     "Export blockchain into file",
		ArgsUsage: "<filename> [<blockNumFirst> <blockNumLast>]",
		Flags: flags.Merge([]cli.Flag{
			CacheFlag,
			SyncModeFlag,
		}, DatabaseFlags),
		Description: `
Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last block to write. In this mode, the file will be appended
if already existing. If the file ends with .gz, the output will
be gzipped.`,
	}
	importHistoryCommand = &cli.Command{
		Action:    importHistory,
		Name:      "import-history",
		Usage:     "Import an Era archive",
		ArgsUsage: "<dir>",
		Flags: flags.Merge([]cli.Flag{},
			DatabaseFlags,
			NetworkFlags,
		),
		Description: `
The import-history command will import blocks and their corresponding receipts
from Era archives.
`,
	}
	exportHistoryCommand = &cli.Command{
		Action:    exportHistory,
		Name:      "export-history",
		Usage:     "Export blockchain history to Era archives",
		ArgsUsage: "<dir> <first> <last>",
		Flags:     flags.Merge(DatabaseFlags),
		Description: `
The export-history command will export blocks and their corresponding receipts
into Era archives. Eras are typically packaged in steps of 8192 blocks.
`,
	}
	importPreimagesCommand = &cli.Command{
		Action:    importPreimages,
		Name:      "import-preimages",
		Usage:     "Import the preimage database from an RLP stream",
		ArgsUsage: "<datafile>",
		Flags: flags.Merge([]cli.Flag{
			CacheFlag,
			SyncModeFlag,
		}, DatabaseFlags),
		Description: `
The import-preimages command imports hash preimages from an RLP encoded stream.
It's deprecated, please use "ixiosSpark db import" instead.
`,
	}

	dumpCommand = &cli.Command{
		Action:    dump,
		Name:      "dump",
		Usage:     "Dump a specific block from storage",
		ArgsUsage: "[? <blockHash> | <blockNum>]",
		Flags: flags.Merge([]cli.Flag{
			CacheFlag,
			IterativeOutputFlag,
			ExcludeCodeFlag,
			ExcludeStorageFlag,
			IncludeIncompletesFlag,
			StartKeyFlag,
			DumpLimitFlag,
		}, DatabaseFlags),
		Description: `
This command dumps out the state for a given block (or latest, if none provided).
`,
	}
)

// initGenesis
func initGenesis(ctx *cli.Context) error {
	return errors.New("init genesis from genesis.json is not supported")
}

func dumpGenesis(ctx *cli.Context) error {
	// check if there is a testnet preset enabled
	var genesis *core.Genesis
	if IsNetworkPreset(ctx) {
		genesis = MakeGenesis(ctx)
	} else if ctx.IsSet(DeveloperFlag.Name) && !ctx.IsSet(DataDirFlag.Name) {
		genesis = core.DeveloperGenesisBlock(11_500_000, nil)
	}

	if genesis != nil {
		if err := json.NewEncoder(os.Stdout).Encode(genesis); err != nil {
			Fatalf("could not encode genesis: %s", err)
		}
		return nil
	}

	// dump whatever already exists in the datadir
	stack, _ := makeConfigNode(ctx)
	for _, name := range []string{"chaindata", "lightchaindata"} {
		db, err := stack.OpenDatabase(name, 0, 0, "", true)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			continue
		}
		genesis, err := core.ReadGenesis(db)
		if err != nil {
			Fatalf("failed to read genesis: %s", err)
		}
		db.Close()

		if err := json.NewEncoder(os.Stdout).Encode(*genesis); err != nil {
			Fatalf("could not encode stored genesis: %s", err)
		}
		return nil
	}
	if ctx.IsSet(DataDirFlag.Name) {
		Fatalf("no existing datadir at %s", stack.Config().DataDir)
	}
	Fatalf("no network preset provided, and no genesis exists in the default datadir")
	return nil
}

func importChain(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		Fatalf("This command requires an argument.")
	}

	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	chain, db := MakeChain(ctx, stack, false)
	defer db.Close()

	// Start periodically gathering memory profiles
	var peakMemAlloc, peakMemSys atomic.Uint64
	go func() {
		stats := new(runtime.MemStats)
		for {
			runtime.ReadMemStats(stats)
			if peakMemAlloc.Load() < stats.Alloc {
				peakMemAlloc.Store(stats.Alloc)
			}
			if peakMemSys.Load() < stats.Sys {
				peakMemSys.Store(stats.Sys)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	// Import the chain
	start := time.Now()

	var importErr error

	if ctx.Args().Len() == 1 {
		if err := ImportChain(chain, ctx.Args().First()); err != nil {
			importErr = err
			log.Error("Import error", "err", err)
		}
	} else {
		for _, arg := range ctx.Args().Slice() {
			if err := ImportChain(chain, arg); err != nil {
				importErr = err
				log.Error("Import error", "file", arg, "err", err)
			}
		}
	}
	chain.Stop()
	fmt.Printf("Import done in %v.\n\n", time.Since(start))

	// Output pre-compaction stats mostly to see the import trashing
	showLeveldbStats(db)

	// Print the memory statistics used by the importing
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	fmt.Printf("Object memory: %.3f MB current, %.3f MB peak\n", float64(mem.Alloc)/1024/1024, float64(peakMemAlloc.Load())/1024/1024)
	fmt.Printf("System memory: %.3f MB current, %.3f MB peak\n", float64(mem.Sys)/1024/1024, float64(peakMemSys.Load())/1024/1024)
	fmt.Printf("Allocations:   %.3f million\n", float64(mem.Mallocs)/1000000)
	fmt.Printf("GC pause:      %v\n\n", time.Duration(mem.PauseTotalNs))

	if ctx.Bool(NoCompactionFlag.Name) {
		return nil
	}

	// Compact the entire database to more accurately measure disk io and print the stats
	start = time.Now()
	fmt.Println("Compacting entire database...")
	if err := db.Compact(nil, nil); err != nil {
		Fatalf("Compaction failed: %v", err)
	}
	fmt.Printf("Compaction done in %v.\n\n", time.Since(start))

	showLeveldbStats(db)
	return importErr
}

func exportChain(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		Fatalf("This command requires an argument.")
	}

	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	chain, db := MakeChain(ctx, stack, true)
	defer db.Close()
	start := time.Now()

	var err error
	fp := ctx.Args().First()
	if ctx.Args().Len() < 3 {
		err = ExportChain(chain, fp)
	} else {
		// This can be improved to allow for numbers larger than 9223372036854775807
		first, ferr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
		last, lerr := strconv.ParseInt(ctx.Args().Get(2), 10, 64)
		if ferr != nil || lerr != nil {
			Fatalf("Export error in parsing parameters: block number not an integer\n")
		}
		if first < 0 || last < 0 {
			Fatalf("Export error: block number must be greater than 0\n")
		}
		if head := chain.CurrentSnapBlock(); uint64(last) > head.Number.Uint64() {
			Fatalf("Export error: block number %d larger than head block %d\n", uint64(last), head.Number.Uint64())
		}
		err = ExportAppendChain(chain, fp, uint64(first), uint64(last))
	}
	if err != nil {
		Fatalf("Export error: %v\n", err)
	}
	fmt.Printf("Export done in %v\n", time.Since(start))
	return nil
}

func importHistory(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		Fatalf("usage: %s", ctx.Command.ArgsUsage)
	}

	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	chain, db := MakeChain(ctx, stack, false)
	defer db.Close()

	var (
		start   = time.Now()
		dir     = ctx.Args().Get(0)
		network string
	)

	// Determine network.
	if IsNetworkPreset(ctx) {
		switch {
		case ctx.Bool(MainnetFlag.Name):
			network = "mainnet"
		case ctx.Bool(NeoDawnFlag.Name):
			network = "neodawn"
		case ctx.Bool(AetherForgeFlag.Name):
			network = "aetherForge"
		case ctx.Bool(AetherBloomFlag.Name):
			network = "aetherBloom"
		}
	} else {
		// No network flag set, try to determine network based on files
		// present in directory.
		var networks []string
		for _, n := range params.NetworkNames {
			entries, err := era.ReadDir(dir, n)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", dir, err)
			}
			if len(entries) > 0 {
				networks = append(networks, n)
			}
		}
		if len(networks) == 0 {
			return fmt.Errorf("no era1 files found in %s", dir)
		}
		if len(networks) > 1 {
			return fmt.Errorf("multiple networks found, use a network flag to specify desired network")
		}
		network = networks[0]
	}

	if err := ImportHistory(chain, db, dir, network); err != nil {
		return err
	}
	fmt.Printf("Import done in %v\n", time.Since(start))
	return nil
}

// exportHistory exports chain history in Era archives at a specified
// directory.
func exportHistory(ctx *cli.Context) error {
	if ctx.Args().Len() != 3 {
		Fatalf("usage: %s", ctx.Command.ArgsUsage)
	}

	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	chain, _ := MakeChain(ctx, stack, true)
	start := time.Now()

	var (
		dir         = ctx.Args().Get(0)
		first, ferr = strconv.ParseInt(ctx.Args().Get(1), 10, 64)
		last, lerr  = strconv.ParseInt(ctx.Args().Get(2), 10, 64)
	)
	if ferr != nil || lerr != nil {
		Fatalf("Export error in parsing parameters: block number not an integer\n")
	}
	if first < 0 || last < 0 {
		Fatalf("Export error: block number must be greater than 0\n")
	}
	if head := chain.CurrentSnapBlock(); uint64(last) > head.Number.Uint64() {
		Fatalf("Export error: block number %d larger than head block %d\n", uint64(last), head.Number.Uint64())
	}
	err := ExportHistory(chain, dir, uint64(first), uint64(last), uint64(era.MaxEra1Size))
	if err != nil {
		Fatalf("Export error: %v\n", err)
	}
	fmt.Printf("Export done in %v\n", time.Since(start))
	return nil
}

// importPreimages imports preimage data from the specified file.
// it is deprecated, and the export function has been removed, but
// the import function is kept around for the time being so that
// older file formats can still be imported.
func importPreimages(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		Fatalf("This command requires an argument.")
	}

	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	db := MakeChainDatabase(ctx, stack, false)
	defer db.Close()
	start := time.Now()

	if err := ImportPreimages(db, ctx.Args().First()); err != nil {
		Fatalf("Import error: %v\n", err)
	}
	fmt.Printf("Import done in %v\n", time.Since(start))
	return nil
}

func parseDumpConfig(ctx *cli.Context, stack *node.Node) (*state.DumpConfig, kvdb.Database, common.Hash, error) {
	db := MakeChainDatabase(ctx, stack, true)
	defer db.Close()

	var header *types.Header
	if ctx.NArg() > 1 {
		return nil, nil, common.Hash{}, fmt.Errorf("expected 1 argument (number or hash), got %d", ctx.NArg())
	}
	if ctx.NArg() == 1 {
		arg := ctx.Args().First()
		if hashish(arg) {
			hash := common.HexToHash(arg)
			if number := rawdb.ReadHeaderNumber(db, hash); number != nil {
				header = rawdb.ReadHeader(db, hash, *number)
			} else {
				return nil, nil, common.Hash{}, fmt.Errorf("block %x not found", hash)
			}
		} else {
			number, err := strconv.ParseUint(arg, 10, 64)
			if err != nil {
				return nil, nil, common.Hash{}, err
			}
			if hash := rawdb.ReadCanonicalHash(db, number); hash != (common.Hash{}) {
				header = rawdb.ReadHeader(db, hash, number)
			} else {
				return nil, nil, common.Hash{}, fmt.Errorf("header for block %d not found", number)
			}
		}
	} else {
		// Use latest
		header = rawdb.ReadHeadHeader(db)
	}
	if header == nil {
		return nil, nil, common.Hash{}, errors.New("no head block found")
	}
	startArg := common.FromHex(ctx.String(StartKeyFlag.Name))
	var start common.Hash
	switch len(startArg) {
	case 0: // common.Hash
	case 32:
		start = common.BytesToHash(startArg)
	case 20:
		start = crypto.Keccak256Hash(startArg)
		log.Info("Converting start-address to hash", "address", common.BytesToAddress(startArg), "hash", start.Hex())
	default:
		return nil, nil, common.Hash{}, fmt.Errorf("invalid start argument: %x. 20 or 32 hex-encoded bytes required", startArg)
	}
	var conf = &state.DumpConfig{
		SkipCode:          ctx.Bool(ExcludeCodeFlag.Name),
		SkipStorage:       ctx.Bool(ExcludeStorageFlag.Name),
		OnlyWithAddresses: !ctx.Bool(IncludeIncompletesFlag.Name),
		Start:             start.Bytes(),
		Max:               ctx.Uint64(DumpLimitFlag.Name),
	}
	log.Info("State dump configured", "block", header.Number, "hash", header.Hash().Hex(),
		"skipcode", conf.SkipCode, "skipstorage", conf.SkipStorage,
		"start", hexutil.Encode(conf.Start), "limit", conf.Max)
	return conf, db, header.Root, nil
}

func dump(ctx *cli.Context) error {
	stack, _ := makeConfigNode(ctx)
	defer stack.Close()

	conf, db, root, err := parseDumpConfig(ctx, stack)
	if err != nil {
		return err
	}
	triedb := MakeTrieDatabase(ctx, db, true, true, false) // always enable preimage lookup
	defer triedb.Close()

	state, err := state.New(root, state.NewDatabaseWithNodeDB(db, triedb), nil)
	if err != nil {
		return err
	}
	if ctx.Bool(IterativeOutputFlag.Name) {
		state.IterativeDump(conf, json.NewEncoder(os.Stdout))
	} else {
		fmt.Println(string(state.Dump(conf)))
	}
	return nil
}

// hashish returns true for strings that look like hashes.
func hashish(x string) bool {
	_, err := strconv.Atoi(x)
	return err != nil
}
