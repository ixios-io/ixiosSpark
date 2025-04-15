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

package ixios

import (
	"math/big"
	"time"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
)

// MinerAPI provides an API to control the sealer.
type MinerAPI struct {
	e *Ixios
}

// NewMinerAPI creates a new MinerAPI instance.
func NewMinerAPI(e *Ixios) *MinerAPI {
	return &MinerAPI{e}
}

// Start starts the sealer with the given number of threads. If threads is nil,
// the number of workers started is equal to the number of logical CPUs that are
// usable by this process. If mining is already running, this method adjust the
// number of threads allowed to use and updates the minimum price required by the
// transaction pool.
func (api *MinerAPI) Start() error {
	return api.e.StartMining()
}

// Stop terminates the sealer, both at the consensus engine level as well as at
// the block creation level.
func (api *MinerAPI) Stop() {
	api.e.StopMining()
}

// SetExtra sets the extra data string that is included when this sealer mines a block.
func (api *MinerAPI) SetExtra(extra string) (bool, error) {
	if err := api.e.Miner().SetExtra([]byte(extra)); err != nil {
		return false, err
	}
	return true, nil
}

// SetGasPrice sets the minimum accepted gas price for the sealer.
func (api *MinerAPI) SetGasPrice(gasPrice hexutil.Big) bool {
	api.e.lock.Lock()
	api.e.gasPrice = (*big.Int)(&gasPrice)
	api.e.lock.Unlock()

	api.e.txPool.SetGasTip((*big.Int)(&gasPrice))
	api.e.Miner().SetGasTip((*big.Int)(&gasPrice))
	return true
}

// SetGasLimit sets the gaslimit to target towards during mining.
func (api *MinerAPI) SetGasLimit(gasLimit hexutil.Uint64) bool {
	api.e.Miner().SetGasCeil(uint64(gasLimit))
	return true
}

// SetEtherbase sets the etherbase of the sealer.
func (api *MinerAPI) SetEtherbase(etherbase common.Address) bool {
	api.e.SetEtherbase(etherbase)
	return true
}

// SetRecommitInterval updates the interval for sealer sealing work recommitting.
func (api *MinerAPI) SetRecommitInterval(interval int) {
	api.e.Miner().SetRecommitInterval(time.Duration(interval) * time.Millisecond)
}
