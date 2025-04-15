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

//go:build darwin || dragonfly || freebsd || linux || nacl || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/ixios-io/ixiosSpark/log"
)

const (
	// On Linux, sun_path is 108 bytes in size
	// see http://man7.org/linux/man-pages/man7/unix.7.html
	maxPathSize = int(108)
)

// ipcListen will create a Unix socket on the given endpoint.
func ipcListen(endpoint string) (net.Listener, error) {
	// account for null-terminator too
	if len(endpoint)+1 > maxPathSize {
		log.Warn(fmt.Sprintf("The ipc endpoint is longer than %d characters. ", maxPathSize-1),
			"endpoint", endpoint)
	}

	// Ensure the IPC path exists and remove any previous leftover
	if err := os.MkdirAll(filepath.Dir(endpoint), 0751); err != nil {
		return nil, err
	}
	err := os.Remove(endpoint)
	l, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(endpoint, 0600)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Listening via IPC"),
		"endpoint", endpoint)
	return l, nil
}

// newIPCConnection will connect to a Unix socket on the given endpoint.
func newIPCConnection(ctx context.Context, endpoint string) (net.Conn, error) {
	fmt.Printf("Connecting to IPC endpoint %s\n", endpoint)
	return new(net.Dialer).DialContext(ctx, "unix", endpoint)
}
