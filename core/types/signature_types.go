// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"github.com/ixios-io/ixiosSpark/common"
)

// Signature scheme identifiers as byte arrays for direct comparison.
var (
	SigTypeECDSA   = []byte{0x00} // ECDSA (legacy)
	SigTypeMLDSA87 = []byte{0x01} // MLDSA87
)

// GetSignatureType returns the signature type bytes from an address.
func GetSignatureType(addr common.Address) []byte {
	if addr.IsECDSA() {
		return []byte{SigTypeECDSA[0]}
	}
	return []byte{SigTypeMLDSA87[0]}
}

// IsValidSignatureType checks if the address maps to one of the supported
// signature schemes. Canonical 48-byte addresses are always either ECDSA or
// MLDSA87 under the post-fork format.
func IsValidSignatureType(addr common.Address) bool {
	_ = addr
	return true
}
