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
	"fmt"

	"github.com/ixios-io/ixiosSpark/ixios"

	// "go.opentelemetry.io/otel"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ixios-io/ixiosSpark/accounts"
	"github.com/ixios-io/ixiosSpark/accounts/keystore"
	"github.com/ixios-io/ixiosSpark/client"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/console/prompt"
	"github.com/ixios-io/ixiosSpark/internal/debug"
	"github.com/ixios-io/ixiosSpark/internal/ethapi"
	"github.com/ixios-io/ixiosSpark/internal/flags"
	"github.com/ixios-io/ixiosSpark/ixios/downloader"
	"github.com/ixios-io/ixiosSpark/log"
	"github.com/ixios-io/ixiosSpark/node"
	"go.uber.org/automaxprocs/maxprocs"

	// OpenTelemetry imports
	//"go.opentelemetry.io/otel"
	//"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	//"go.opentelemetry.io/otel/propagation"
	//"go.opentelemetry.io/otel/sdk/resource"
	//"go.opentelemetry.io/otel/sdk/trace"
	//semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	//oteltrace "go.opentelemetry.io/otel/trace"

	// Force-load the tracer engines to trigger registration
	_ "github.com/ixios-io/ixiosSpark/ixios/tracers/js"
	_ "github.com/ixios-io/ixiosSpark/ixios/tracers/native"

	"github.com/urfave/cli/v2"
)

const (
	clientIdentifier = "ixiosSpark" // Client identifier to advertise over the network
)

var (
	// flags that configure the node
	nodeFlags = flags.Merge([]cli.Flag{
		IdentityFlag,
		UnlockedAccountFlag,
		PasswordFileFlag,
		EnableBroadcastFlag,
		BootnodesFlag,
		MinFreeDiskSpaceFlag,
		KeyStoreDirFlag,
		ExternalSignerFlag,
		USBFlag,
		OverrideCancun,
		TxPoolLocalsFlag,
		TxPoolNoLocalsFlag,
		TxPoolJournalFlag,
		TxPoolRejournalFlag,
		TxPoolPriceLimitFlag,
		TxPoolPriceBumpFlag,
		TxPoolAccountSlotsFlag,
		TxPoolGlobalSlotsFlag,
		TxPoolAccountQueueFlag,
		TxPoolGlobalQueueFlag,
		TxPoolLifetimeFlag,
		BlobPoolDataDirFlag,
		BlobPoolDataCapFlag,
		BlobPoolPriceBumpFlag,
		SyncModeFlag,
		SyncTargetFlag,
		ExitWhenSyncedFlag,
		GCModeFlag,
		SnapshotFlag,
		TransactionHistoryFlag,
		StateHistoryFlag,
		EthRequiredBlocksFlag,
		BloomFilterSizeFlag,
		CacheFlag,
		CacheDatabaseFlag,
		CacheTrieFlag,
		CacheGCFlag,
		CacheSnapshotFlag,
		CacheNoPrefetchFlag,
		CachePreimagesFlag,
		CacheLogSizeFlag,
		FDLimitFlag,
		CryptoKZGFlag,
		ListenPortFlag,
		DiscoveryPortFlag,
		MaxPeersFlag,
		MaxPendingPeersFlag,
		MiningEnabledFlag,
		MinerGasLimitFlag,
		MinerGasPriceFlag,
		MinerEtherbaseFlag,
		MinerExtraDataFlag,
		MinerRecommitIntervalFlag,
		MinerNewPayloadTimeout,
		NATFlag,
		NoDiscoverFlag,
		DiscoveryV4Flag,
		DiscoveryV5Flag,
		NetrestrictFlag,
		NodeKeyFileFlag,
		NodeKeyHexFlag,
		DNSDiscoveryFlag,
		DeveloperFlag,
		DeveloperGasLimitFlag,
		DeveloperPeriodFlag,
		VMEnableDebugFlag,
		NetworkIdFlag,
		NoCompactionFlag,
		GpoBlocksFlag,
		GpoPercentileFlag,
		GpoMaxGasPriceFlag,
		GpoIgnoreGasPriceFlag,
		configFileFlag,
	}, NetworkFlags, DatabaseFlags)

	rpcFlags = []cli.Flag{
		HTTPEnabledFlag,
		HTTPListenAddrFlag,
		HTTPPortFlag,
		HTTPCORSDomainFlag,
		AuthListenFlag,
		AuthPortFlag,
		AuthVirtualHostsFlag,
		JWTSecretFlag,
		HTTPVirtualHostsFlag,
		HTTPApiFlag,
		HTTPPathPrefixFlag,
		WSEnabledFlag,
		WSListenAddrFlag,
		WSPortFlag,
		WSApiFlag,
		WSAllowedOriginsFlag,
		WSPathPrefixFlag,
		IPCDisabledFlag,
		IPCPathFlag,
		InsecureUnlockAllowedFlag,
		RPCGlobalGasCapFlag,
		RPCGlobalEVMTimeoutFlag,
		RPCGlobalTxFeeCapFlag,
		AllowUnprotectedTxs,
		BatchRequestLimit,
		BatchResponseMaxSize,
	}
)

var app = flags.NewApp("IxiosSpark command line interface")

func init() {
	// Initialize the CLI app and start
	app.Action = ixiosCLI
	app.Commands = []*cli.Command{
		// See chaincmd.go:
		importCommand,
		exportCommand,
		importHistoryCommand,
		exportHistoryCommand,
		importPreimagesCommand,
		removedbCommand,
		dumpCommand,
		dumpGenesisCommand,
		// See accountcmd.go:
		accountCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		licenseCommand,
		// See config.go
		dumpConfigCommand,
		// see dbcmd.go
		dbCommand,
		// See snapshot.go
		snapshotCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = flags.Merge(
		nodeFlags,
		rpcFlags,
		consoleFlags,
		debug.Flags,
	)
	flags.AutoEnvVars(app.Flags, "IXIOS")

	app.Before = func(ctx *cli.Context) error {
		maxprocs.Set() // Automatically set GOMAXPROCS to match Linux container CPU quota.
		flags.MigrateGlobalFlags(ctx)
		if err := debug.Setup(ctx); err != nil {
			return err
		}
		flags.CheckEnvVars(ctx, app.Flags, "IXIOS")
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		prompt.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// prepare manipulates memory cache allowance and setups metric system.
// This function should be called before launching devp2p stack.
func prepare(ctx *cli.Context) {
	// If we're running a known preset, log it for convenience.
	switch {
	case ctx.IsSet(AetherForgeFlag.Name):
		log.Info("Starting ixiosSpark on AetherForge testnet...")

	case ctx.IsSet(NeoDawnFlag.Name):
		log.Info("Starting ixiosSpark on NeoDawn testnet...")

	case ctx.IsSet(DeveloperFlag.Name):
		log.Info("Starting ixiosSpark in ephemeral dev mode...")
		log.Warn(`You are running ixiosSpark in --dev mode. Please note the following:

  1. This mode is only intended for fast, iterative development without assumptions on
     security or persistence.
  2. The database is created in memory unless specified otherwise. Therefore, shutting down
     your computer or losing power will wipe your entire block data and chain state for
     your dev environment.
  3. A random, pre-allocated developer account will be available and unlocked as
     eth.coinbase, which can be used for testing. The random dev account is temporary,
     stored on a ramdisk, and will be lost if your machine is restarted.
  4. Mining is enabled by default. However, the client will only seal blocks if transactions
     are pending in the mempool. The sealer's minimum accepted gas price is 1.
  5. Networking is disabled; there is no listen-address, the maximum number of peers is set
     to 0, and discovery is disabled.
`)

	case !ctx.IsSet(NetworkIdFlag.Name):
		// Make sure we're not on any supported preconfigured testnet either
		if !ctx.IsSet(NeoDawnFlag.Name) &&
			!ctx.IsSet(AetherForgeFlag.Name) &&
			!ctx.IsSet(AetherBloomFlag.Name) &&
			!ctx.IsSet(DeveloperFlag.Name) {
			// We're really on mainnet.
			log.Info("Starting on Ixios L1 mainnet...")
		}
	}
	// If we're a full node on mainnet without --cache specified, bump default cache allowance
	if !ctx.IsSet(CacheFlag.Name) && !ctx.IsSet(NetworkIdFlag.Name) {
		// Make sure we're not on any supported preconfigured testnet either
		if !ctx.IsSet(NeoDawnFlag.Name) &&
			!ctx.IsSet(AetherForgeFlag.Name) &&
			!ctx.IsSet(AetherBloomFlag.Name) &&
			!ctx.IsSet(DeveloperFlag.Name) {
			// Nope, we're really on mainnet. Bump that cache up!
			log.Info("Bumping default cache on mainnet", "provided", ctx.Int(CacheFlag.Name), "updated", 4096)
			ctx.Set(CacheFlag.Name, strconv.Itoa(4096))
		}
	}
}

// ixiosCLI is the main entry point into the system if no special subcommand is run.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func ixiosCLI(ctx *cli.Context) error {
	if args := ctx.Args().Slice(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	prepare(ctx)
	stack, backend := makeFullNode(ctx)
	defer stack.Close()

	// OpenTelemetry
	// implement here

	startNode(ctx, stack, backend, false)
	stack.Wait()
	return nil

	/*
		// Start a root context for our OpenTelemetry instrumentation
		rootCtx := context.Background()

		// Set up OpenTelemetry, returning a shutdown function we can defer
		otelShutdown, err := setupOpenTelemetry(rootCtx)
		if err != nil {
			return fmt.Errorf("failed to set up OpenTelemetry: %w", err)
		}
		defer func() {
			// Make sure we shut down cleanly, flushing any remaining spans
			shutdownErr := otelShutdown(context.Background())
			if shutdownErr != nil {
				log.Error("Error shutting down OpenTelemetry", "err", shutdownErr)
			}
		}()

		// Start a span for the overall CLI execution
		tracer := otel.Tracer("github.com/ixios-io/ixiosSpark/main")
		cliCtx, span := tracer.Start(rootCtx, "ixiosCLI")
		defer span.End()

		// If there are any leftover args not recognized by the CLI, return an error
		if args := ctx.Args().Slice(); len(args) > 0 {
			return fmt.Errorf("invalid command: %q", args[0])
		}

		prepare(ctx)

		// Construct the full node
		stack, backend := makeFullNode(ctx)
		defer stack.Close()

		// Boot up the node
		startNode(ctx, stack, backend, false)
		stack.Wait()

		// End the top-level CLI span, ensuring instrumentation for the entire CLI run is captured
		return nil
	*/
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// sealer.
func startNode(ctx *cli.Context, stack *node.Node, backend ethapi.Backend, isConsole bool) {
	// Start up the node itself
	StartNode(ctx, stack, isConsole)

	// Unlock any account specifically requested
	unlockAccounts(ctx, stack)

	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	// Create a client to interact with local ixiosSpark node.
	rpcClient := stack.Attach()
	ethClient := client.NewClient(rpcClient)

	go func() {
		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				var derivationPaths []accounts.DerivationPath
				if event.Wallet.URL().Scheme == "ledger" {
					derivationPaths = append(derivationPaths, accounts.LegacyLedgerBaseDerivationPath)
				}
				derivationPaths = append(derivationPaths, accounts.DefaultBaseDerivationPath)

				event.Wallet.SelfDerive(derivationPaths, ethClient)

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()

	// Spawn a standalone goroutine for status synchronization monitoring,
	// close the node when synchronization is complete if user required.
	if ctx.Bool(ExitWhenSyncedFlag.Name) {
		go func() {
			sub := stack.EventMux().Subscribe(downloader.DoneEvent{})
			defer sub.Unsubscribe()
			for {
				event := <-sub.Chan()
				if event == nil {
					continue
				}
				done, ok := event.Data.(downloader.DoneEvent)
				if !ok {
					continue
				}
				if timestamp := time.Unix(int64(done.Latest.Time), 0); time.Since(timestamp) < 10*time.Minute {
					log.Info("Synchronisation completed", "latestnum", done.Latest.Number, "latesthash", done.Latest.Hash(),
						"age", common.PrettyAge(timestamp))
					stack.Close()
				}
			}
		}()
	}

	// Start auxiliary services if enabled
	if ctx.Bool(MiningEnabledFlag.Name) {
		// Mining only makes sense if a full Ixios node is running
		if ctx.String(SyncModeFlag.Name) == "light" {
			Fatalf("Light clients do not support mining")
		}
		ethBackend, ok := backend.(*ixios.EthAPIBackend)
		if !ok {
			Fatalf("Ixios L1 service not running")
		}
		// Set the gas price to the limits from the CLI and start mining
		gasprice := flags.GlobalBig(ctx, MinerGasPriceFlag.Name)
		ethBackend.TxPool().SetGasTip(gasprice)
		if err := ethBackend.StartMining(); err != nil {
			Fatalf("Failed to start mining: %v", err)
		}
	}
}

// unlockAccounts unlocks any account specifically requested.
func unlockAccounts(ctx *cli.Context, stack *node.Node) {
	var unlocks []string
	inputs := strings.Split(ctx.String(UnlockedAccountFlag.Name), ",")
	for _, input := range inputs {
		if trimmed := strings.TrimSpace(input); trimmed != "" {
			unlocks = append(unlocks, trimmed)
		}
	}
	// Short circuit if there is no account to unlock.
	if len(unlocks) == 0 {
		return
	}
	// If insecure account unlocking is not allowed if node's APIs are exposed to external.
	// Print warning log to user and skip unlocking.
	if !stack.Config().InsecureUnlockAllowed && stack.Config().ExtRPCEnabled() {
		Fatalf("Account unlock with HTTP access is forbidden!")
	}
	backends := stack.AccountManager().Backends(keystore.KeyStoreType)
	if len(backends) == 0 {
		log.Warn("Failed to unlock accounts, keystore is not available")
		return
	}
	ks := backends[0].(*keystore.KeyStore)
	passwords := MakePasswordList(ctx)
	for i, account := range unlocks {
		unlockAccount(ks, account, i, passwords)
	}
}
