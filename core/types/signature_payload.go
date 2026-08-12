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
	"errors"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/ixios-io/ixiosSpark/common"
)

var (
	ErrUnsupportedTxSignatureType         = errors.New("unsupported transaction signature type")
	ErrInvalidTransactionSignaturePayload = errors.New("invalid transaction signature payload")
	ErrMLDSA87RequiresProtectedOrTypedTx  = errors.New("MLDSA87 signatures require EIP-155 or typed transactions")
)

type signaturePayloadAccessor interface {
	signatureType() []byte
	signaturePublicKey() []byte
	signatureBytes() []byte
	setSignaturePayload(sigType, publicKey, signature []byte)
}

func signatureTypeField(inner TxData) []byte {
	if payloadTx, ok := inner.(signaturePayloadAccessor); ok {
		return payloadTx.signatureType()
	}
	return nil
}

func signaturePublicKeyField(inner TxData) []byte {
	if payloadTx, ok := inner.(signaturePayloadAccessor); ok {
		return payloadTx.signaturePublicKey()
	}
	return nil
}

func signatureBytesField(inner TxData) []byte {
	if payloadTx, ok := inner.(signaturePayloadAccessor); ok {
		return payloadTx.signatureBytes()
	}
	return nil
}

func setSignaturePayload(inner TxData, sigType, publicKey, signature []byte) {
	if payloadTx, ok := inner.(signaturePayloadAccessor); ok {
		payloadTx.setSignaturePayload(sigType, publicKey, signature)
	}
}

func clearSignaturePayload(inner TxData) {
	setSignaturePayload(inner, nil, nil, nil)
}

func signaturePayloadPresent(sigType, publicKey, signature []byte) bool {
	return len(sigType) != 0 || len(publicKey) != 0 || len(signature) != 0
}

func hasMLDSA87SignaturePayload(tx *Transaction) bool {
	sigType := signatureTypeField(tx.inner)
	publicKey := signaturePublicKeyField(tx.inner)
	signature := signatureBytesField(tx.inner)
	if bytes.Equal(sigType, SigTypeMLDSA87) {
		return true
	}
	return len(publicKey) != 0 || len(signature) != 0
}

func validateSignaturePayload(inner TxData) error {
	return validateSignaturePayloadRaw(
		signatureTypeField(inner),
		signaturePublicKeyField(inner),
		signatureBytesField(inner),
	)
}

func validateSignaturePayloadRaw(sigType, publicKey, signature []byte) error {
	if !signaturePayloadPresent(sigType, publicKey, signature) {
		return nil
	}
	if !bytes.Equal(sigType, SigTypeMLDSA87) {
		return ErrUnsupportedTxSignatureType
	}
	if len(publicKey) != mldsa87.PublicKeySize || len(signature) != mldsa87.SignatureSize {
		return ErrInvalidTransactionSignaturePayload
	}
	return nil
}

// SignatureType returns the transaction signature type.
// Legacy transactions without an explicit post-quantum payload remain ECDSA.
func (tx *Transaction) SignatureType() []byte {
	if hasMLDSA87SignaturePayload(tx) {
		return common.CopyBytes(SigTypeMLDSA87)
	}
	return common.CopyBytes(SigTypeECDSA)
}

// SignaturePublicKey returns the explicit transaction signature public key payload, if any.
func (tx *Transaction) SignaturePublicKey() []byte {
	return common.CopyBytes(signaturePublicKeyField(tx.inner))
}

// SignatureBytes returns the explicit transaction signature payload, if any.
func (tx *Transaction) SignatureBytes() []byte {
	return common.CopyBytes(signatureBytesField(tx.inner))
}
