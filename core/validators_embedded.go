// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// Copyright 2025 The IxiosSpark Authors
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/crypto"
)

const pubKeyHexSize = 132 // e.g., "0x" + 130 hex chars => 132 ASCII bytes
const mksDataSize = 2656
const readChunkSize = pubKeyHexSize + mksDataSize // = 2,788
const chunkSize = mksDataSize + common.AddressLength

type ValidatorMKS struct {
	Address common.Address // 32 bytes: first 6 bytes are zero, last 26 bytes are keccak(...)
	MKSData []byte         // 5,348 bytes
}

// validatorsDat is the raw binary data with all validators combined.
//
//go:embed data/validators.dat
var validatorsDat []byte

// parseEmbeddedValidators splits validatorsDat into a slice of ValidatorMKS objects.
//
// The on-disk format is:
//   - [0..131]: ASCII-hex ECDSA public key (including leading "0x"), total 132 bytes
//   - [132..2788]: raw binary MKS data
//
// The 132-byte hex-string public key is converted into a 65-byte uncompressed public key,
// then hashed (Keccak256). The Ixios address is 32 bytes: for legacy ECDSA-26 address format,
// the first 6 bytes are 0x00 and the remaining 26 bytes are Keccak256(pubKey[1:])
//
// This function returns an error if the total length of validatorsDat is not
// a clean multiple of chunkSize, or if any chunk fails hex/public-key validation.
func parseEmbeddedValidators() ([]ValidatorMKS, error) {
	if len(validatorsDat) == 0 {
		return nil, errors.New("validator data is empty")
	}
	if len(validatorsDat)%readChunkSize != 0 {
		return nil, fmt.Errorf(
			"validator data length %d is not a multiple of chunkSize %d",
			len(validatorsDat), readChunkSize,
		)
	}

	numChunks := len(validatorsDat) / readChunkSize
	results := make([]ValidatorMKS, 0, numChunks)

	for i := 0; i < numChunks; i++ {
		start := i * readChunkSize
		end := start + readChunkSize
		chunk := validatorsDat[start:end]

		pubKeyHex := string(chunk[:pubKeyHexSize])
		if len(pubKeyHex) != pubKeyHexSize || pubKeyHex[0:2] != "0x" {
			return nil, fmt.Errorf("invalid hex pubkey format at chunk %d", i)
		}

		pubKeyBytes, err := hex.DecodeString(pubKeyHex[2:])
		if err != nil {
			return nil, fmt.Errorf("chunk %d hex decode error: %v", i, err)
		}
		if len(pubKeyBytes) != 65 {
			return nil, fmt.Errorf("chunk %d public key is %d bytes, expected 65", i, len(pubKeyBytes))
		}

		// Derive the Ixios address
		fullHash := crypto.Keccak256(pubKeyBytes[1:])
		if len(fullHash) != 32 {
			return nil, fmt.Errorf("chunk %d keccak result length mismatch", i)
		}
		var ixiosAddress common.Address // 32 bytes
		copy(ixiosAddress[6:], fullHash[len(fullHash)-26:])

		// Derive the MKS data
		mks := make([]byte, mksDataSize)
		copy(mks, chunk[pubKeyHexSize:])

		results = append(results, ValidatorMKS{
			Address: ixiosAddress,
			MKSData: mks,
		})
	}

	return results, nil
}
