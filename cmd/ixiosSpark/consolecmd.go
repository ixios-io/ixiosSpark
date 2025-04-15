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
	"time"

	"github.com/ixios-io/ixiosSpark/console"
	"github.com/ixios-io/ixiosSpark/internal/flags"
	"github.com/urfave/cli/v2"
)

var (
	consoleFlags = []cli.Flag{JSpathFlag, ExecFlag, PreloadJSFlag}

	consoleCommand = &cli.Command{
		Action: localConsole,
		Name:   "console",
		Usage:  "Start an interactive JavaScript environment",
		Flags:  flags.Merge(nodeFlags, rpcFlags, consoleFlags),
		Description: `
The Ixios console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.`,
	}

	attachCommand = &cli.Command{
		Action:    remoteConsole,
		Name:      "attach",
		Usage:     "Start an interactive JavaScript environment (connect to node)",
		ArgsUsage: "[endpoint]",
		Flags:     flags.Merge([]cli.Flag{DataDirFlag, HttpHeaderFlag}, consoleFlags),
		Description: `
The Ixios console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
This command allows to open a console on a running ixiosSpark node.`,
	}
)

// localConsole starts a new ixiosSpark node and attaches a JavaScript console to it.
func localConsole(ctx *cli.Context) error {
	// Create and start the node based on the CLI flags
	prepare(ctx)
	stack, backend := makeFullNode(ctx)
	time.Sleep(100 * time.Millisecond)
	go startNode(ctx, stack, backend, true)
	defer stack.Close()

	// Attach to the newly started node and create the JavaScript console.
	client := stack.Attach()
	config := console.Config{
		DataDir: MakeDataDir(ctx),
		DocRoot: ctx.String(JSpathFlag.Name),
		Client:  client,
		Preload: MakeConsolePreloads(ctx),
	}
	console, err := console.New(config)
	if err != nil {
		return fmt.Errorf("failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	// If only a short execution was requested, evaluate and return.
	if script := ctx.String(ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}

	// Track node shutdown and stop the console when it goes down.
	// This happens when SIGTERM is sent to the process.
	go func() {
		stack.Wait()
		console.StopInteractive()
	}()

	// Print the welcome screen and enter interactive mode.
	console.Welcome()
	console.Interactive()
	return nil
}

// remoteConsole will connect to a remote geth instance, attaching a JavaScript
// console to it.
func remoteConsole(ctx *cli.Context) error {
	if ctx.Args().Len() > 1 {
		Fatalf("invalid command-line: too many arguments")
	}
	endpoint := ctx.Args().First()
	if endpoint == "" {
		cfg := defaultNodeConfig()
		SetDataDir(ctx, &cfg)
		endpoint = cfg.IPCEndpoint()
	}
	client, err := DialRPCWithHeaders(endpoint, ctx.StringSlice(HttpHeaderFlag.Name))
	if err != nil {
		Fatalf("Unable to attach to remote ixiosSpark: %v", err)
	}
	config := console.Config{
		DataDir: MakeDataDir(ctx),
		DocRoot: ctx.String(JSpathFlag.Name),
		Client:  client,
		Preload: MakeConsolePreloads(ctx),
	}
	console, err := console.New(config)
	if err != nil {
		Fatalf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	if script := ctx.String(ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}

	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()
	return nil
}
