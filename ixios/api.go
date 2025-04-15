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
	"github.com/ixios-io/ixiosSpark/common"
)

// IxiosAPI provides an API to access Ixios full node-related information.
type IxiosAPI struct {
	e *Ixios
}

// NewIxiosAPI creates a new Ixios protocol API for full nodes.
func NewIxiosAPI(e *Ixios) *IxiosAPI {
	return &IxiosAPI{e}
}

// Etherbase is the address that mining rewards will be sent to.
func (api *IxiosAPI) Etherbase() (common.Address, error) {
	return api.e.Etherbase()
}

// Coinbase is the address that mining rewards will be sent to (alias for Etherbase).
func (api *IxiosAPI) Coinbase() (common.Address, error) {
	return api.Etherbase()
}

// Mining returns an indication if this node is currently mining.
func (api *IxiosAPI) Mining() bool {
	return api.e.IsMining()
}

// Mining returns an indication if this node is currently mining.
func (api *IxiosAPI) Sealing() bool {
	return api.e.IsMining()
}
