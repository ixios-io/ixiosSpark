// Copyright 2024 The IxiosSpark Authors
// This file is part of the IxiosSpark library
//
// The IxiosSpark library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The IxiosSpark library is distributed with the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"github.com/ixios-io/ixiosSpark/common"
)

// SigningScheme represents a specific signature scheme implementation
type SigningScheme interface {
	// Type returns the signature scheme identifier (first 6 bytes of addresses)
	Type() []byte

	// Sign signs the given hash
	Sign(hash []byte, key interface{}) ([]byte, error)

	// Verify verifies a signature against a hash and returns the recovered address
	Verify(hash []byte, sig []byte) (common.Address, error)

	// ValidateKey ensures a private key is valid for this scheme
	ValidateKey(key interface{}) error
}

// SigningKey represents a key that can be used with a specific signature scheme
type SigningKey interface {
	// Scheme returns the SigningScheme this key is for
	Scheme() SigningScheme
}
