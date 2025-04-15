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

package downloader

import "fmt"

// SyncMode represents the synchronisation mode of the downloader.
// It is a uint32 as it is used with atomic operations.
type SyncMode uint32

const (
	FullSync  SyncMode = iota // Synchronise the entire blockchain history from full blocks
	SnapSync                  // Download the chain and the state via compact snapshots
	LightSync                 // Download only the headers and terminate afterwards
)

func (mode SyncMode) IsValid() bool {
	return mode >= FullSync && mode <= LightSync
}

// String implements the stringer interface.
func (mode SyncMode) String() string {
	switch mode {
	case FullSync:
		return "full"
	case SnapSync:
		return "snap"
	case LightSync:
		return "light"
	default:
		return "unknown"
	}
}

func (mode SyncMode) MarshalText() ([]byte, error) {
	switch mode {
	case FullSync:
		return []byte("full"), nil
	case SnapSync:
		return []byte("snap"), nil
	case LightSync:
		return []byte("light"), nil
	default:
		return nil, fmt.Errorf("unknown sync mode %d", mode)
	}
}

func (mode *SyncMode) UnmarshalText(text []byte) error {
	switch string(text) {
	case "full":
		*mode = FullSync
	case "snap":
		*mode = SnapSync
	case "light":
		*mode = LightSync
	default:
		return fmt.Errorf(`unknown sync mode %q, want "full", "snap" or "light"`, text)
	}
	return nil
}
