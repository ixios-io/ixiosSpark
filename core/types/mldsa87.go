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
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/ixios-io/ixiosSpark/common"
)

const mldsa87SignatureEnvelopeSize = 1 + mldsa87.PublicKeySize + mldsa87.SignatureSize

func isMLDSA87SignatureEnvelope(sig []byte) bool {
	return len(sig) == mldsa87SignatureEnvelopeSize && len(sig) > 0 && sig[0] == SigTypeMLDSA87[0]
}

func decodeMLDSA87SignatureEnvelope(sig []byte) (publicKey, signature []byte, err error) {
	if !isMLDSA87SignatureEnvelope(sig) {
		return nil, nil, ErrInvalidTransactionSignaturePayload
	}
	publicKey = common.CopyBytes(sig[1 : 1+mldsa87.PublicKeySize])
	signature = common.CopyBytes(sig[1+mldsa87.PublicKeySize:])
	return publicKey, signature, nil
}

func encodeMLDSA87SignatureEnvelope(publicKey, signature []byte) ([]byte, error) {
	if len(publicKey) != mldsa87.PublicKeySize || len(signature) != mldsa87.SignatureSize {
		return nil, fmt.Errorf("%w: want %d-byte public key and %d-byte signature", ErrInvalidTransactionSignaturePayload, mldsa87.PublicKeySize, mldsa87.SignatureSize)
	}
	encoded := make([]byte, mldsa87SignatureEnvelopeSize)
	encoded[0] = SigTypeMLDSA87[0]
	copy(encoded[1:], publicKey)
	copy(encoded[1+mldsa87.PublicKeySize:], signature)
	return encoded, nil
}

func signMLDSA87Hash(hash common.Hash, prv *mldsa87.PrivateKey) ([]byte, error) {
	signature := make([]byte, mldsa87.SignatureSize)
	if err := mldsa87.SignTo(prv, hash[:], nil, false, signature); err != nil {
		return nil, err
	}
	publicKey, ok := prv.Public().(*mldsa87.PublicKey)
	if !ok {
		return nil, errors.New("invalid MLDSA87 public key")
	}
	return encodeMLDSA87SignatureEnvelope(publicKey.Bytes(), signature)
}

// SignTxMLDSA87 signs the transaction using the given signer and MLDSA87 private key.
func SignTxMLDSA87(tx *Transaction, s Signer, prv *mldsa87.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)
	sig, err := signMLDSA87Hash(h, prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
}

// SignNewTxMLDSA87 creates a transaction and signs it with MLDSA87.
func SignNewTxMLDSA87(prv *mldsa87.PrivateKey, s Signer, txdata TxData) (*Transaction, error) {
	tx := NewTx(txdata)
	return SignTxMLDSA87(tx, s, prv)
}

// MustSignNewTxMLDSA87 creates a transaction and signs it with MLDSA87.
// This panics if the transaction cannot be signed.
func MustSignNewTxMLDSA87(prv *mldsa87.PrivateKey, s Signer, txdata TxData) *Transaction {
	tx, err := SignNewTxMLDSA87(prv, s, txdata)
	if err != nil {
		panic(err)
	}
	return tx
}
