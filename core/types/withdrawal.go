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

package types

import (
	"bytes"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/common/hexutil"
	"github.com/ixios-io/ixiosSpark/rlp"
)

//go:generate go run github.com/fjl/gencodec -type Withdrawal -field-override withdrawalMarshaling -out gen_withdrawal_json.go
//go:generate go run ../../rlp/rlpgen -type Withdrawal -out gen_withdrawal_rlp.go

// Withdrawal represents a validator withdrawal from the consensus layer.
type Withdrawal struct {
	Index     uint64         `json:"index"`          // monotonically increasing identifier issued by consensus layer
	Validator uint64         `json:"validatorIndex"` // index of validator associated with withdrawal
	Address   common.Address `json:"address"`        // target address for withdrawn ether
	Amount    uint64         `json:"amount"`         // value of withdrawal in Gwei
}

// field type overrides for gencodec
type withdrawalMarshaling struct {
	Index     hexutil.Uint64
	Validator hexutil.Uint64
	Amount    hexutil.Uint64
}

// Withdrawals implements DerivableList for withdrawals.
type Withdrawals []*Withdrawal

// Len returns the length of s.
func (s Withdrawals) Len() int { return len(s) }

// EncodeIndex encodes the i'th withdrawal to w. Note that this does not check for errors
// because we assume that *Withdrawal will only ever contain valid withdrawals that were either
// constructed by decoding or via public API in this package.
func (s Withdrawals) EncodeIndex(i int, w *bytes.Buffer) {
	rlp.Encode(w, s[i])
}
