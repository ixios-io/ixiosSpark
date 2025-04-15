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
	"bytes"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/params"
)

// Signature scheme identifiers as byte arrays for direct comparison
var (
	SigTypeECDSA2    = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // ECDSA-2 (legacy)
	SigTypeECDSA26   = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01} // ECDSA-26
	SigTypeDilith2   = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02} // Dilithium-2
	SigTypeDilith3   = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x03} // Dilithium-3
	SigTypeDilith5   = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x05} // Dilithium-5
	SigTypeFalcon512 = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x06} // Falcon512
)

// GetSignatureType returns the signature type bytes from an address
func GetSignatureType(addr common.Address) []byte {
	return addr[:params.SignaturePrefixLength]
}

// IsValidSignatureType checks if the address starts with a valid signature prefix
func IsValidSignatureType(addr common.Address) bool {
	prefix := GetSignatureType(addr)
	return bytes.Equal(prefix, SigTypeECDSA2) ||
		bytes.Equal(prefix, SigTypeECDSA26) ||
		bytes.Equal(prefix, SigTypeDilith2) ||
		bytes.Equal(prefix, SigTypeDilith3) ||
		bytes.Equal(prefix, SigTypeDilith5) ||
		bytes.Equal(prefix, SigTypeFalcon512)
}

// IsDilithiumAddress returns true if the address uses any Dilithium signature scheme
func IsDilithiumAddress(addr common.Address) bool {
	prefix := GetSignatureType(addr)
	return bytes.Equal(prefix, SigTypeDilith2) ||
		bytes.Equal(prefix, SigTypeDilith3) ||
		bytes.Equal(prefix, SigTypeDilith5)
}
