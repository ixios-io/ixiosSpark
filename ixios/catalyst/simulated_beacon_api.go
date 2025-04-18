// Copyright 2023 The go-ethereum Authors
// This file is part of the IxiosSpark library, which builds upon the source code of the go-ethereum library.
//
// The IxiosSpark library, including the go-ethereum library source code it is based on, is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The IxiosSpark and go-ethereum library source code are distributed with the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package catalyst

import (
	"context"
	"time"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core"
	"github.com/ixios-io/ixiosSpark/core/types"
	"github.com/ixios-io/ixiosSpark/log"
)

type api struct {
	sim *SimulatedBeacon
}

func (a *api) loop() {
	var (
		newTxs = make(chan core.NewTxsEvent)
		sub    = a.sim.eth.TxPool().SubscribeTransactions(newTxs, true)
	)
	defer sub.Unsubscribe()

	for {
		select {
		case <-a.sim.shutdownCh:
			return
		case w := <-a.sim.withdrawals.pending:
			withdrawals := append(a.sim.withdrawals.gatherPending(9), w)
			if err := a.sim.sealBlock(withdrawals, uint64(time.Now().Unix())); err != nil {
				log.Warn("Error performing sealing work", "err", err)
			}
		case <-newTxs:
			a.sim.Commit()
		}
	}
}

func (a *api) AddWithdrawal(ctx context.Context, withdrawal *types.Withdrawal) error {
	return a.sim.withdrawals.add(withdrawal)
}

func (a *api) SetFeeRecipient(ctx context.Context, feeRecipient common.Address) {
	a.sim.setFeeRecipient(feeRecipient)
}
