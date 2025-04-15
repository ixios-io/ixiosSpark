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

package blobpool

import (
	"github.com/ixios-io/ixiosSpark/log"
)

// Config are the configuration parameters of the blob transaction pool.
type Config struct {
	Datadir   string // Data directory containing the currently executable blobs
	Datacap   uint64 // Soft-cap of database storage (hard cap is larger due to overhead)
	PriceBump uint64 // Minimum price bump percentage to replace an already existing nonce
}

// DefaultConfig contains the default configurations for the transaction pool.
var DefaultConfig = Config{
	Datadir:   "blobpool",
	Datacap:   10 * 1024 * 1024 * 1024,
	PriceBump: 100, // either have patience or be aggressive, no mushy ground
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *Config) sanitize() Config {
	conf := *config
	if conf.Datacap < 1 {
		log.Warn("Sanitizing invalid blobpool storage cap", "provided", conf.Datacap, "updated", DefaultConfig.Datacap)
		conf.Datacap = DefaultConfig.Datacap
	}
	if conf.PriceBump < 1 {
		log.Warn("Sanitizing invalid blobpool price bump", "provided", conf.PriceBump, "updated", DefaultConfig.PriceBump)
		conf.PriceBump = DefaultConfig.PriceBump
	}
	return conf
}
