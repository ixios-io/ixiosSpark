package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ixios-io/ixiosSpark/common/hexutil"
	"github.com/ixios-io/ixiosSpark/rlp"
)

func hasZeroPrefixBytes(b []byte, prefixLen int) bool {
	if prefixLen > len(b) {
		prefixLen = len(b)
	}
	for i := 0; i < prefixLen; i++ {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

func isAcceptedAddressLength(length int) bool {
	switch length {
	case 20, LegacyAddressLength, AddressLength:
		return true
	default:
		return false
	}
}

func decodeAddressText(input []byte, requirePrefix bool) ([]byte, error) {
	if len(input) == 0 {
		return nil, hexutil.ErrEmptyString
	}
	raw := input
	if len(raw) >= 2 && raw[0] == '0' && (raw[1] == 'x' || raw[1] == 'X') {
		raw = raw[2:]
	} else if requirePrefix {
		return nil, hexutil.ErrMissingPrefix
	}
	if len(raw)%2 != 0 {
		return nil, hexutil.ErrOddLength
	}
	if !isHex(string(raw)) {
		return nil, hexutil.ErrSyntax
	}
	if !isAcceptedAddressLength(len(raw) / 2) {
		return nil, fmt.Errorf("invalid address length %d", len(raw)/2)
	}
	out := make([]byte, len(raw)/2)
	if _, err := hex.Decode(out, raw); err != nil {
		return nil, err
	}
	return out, nil
}

// IsECDSA reports whether the canonical 48-byte address uses the ECDSA layout
// (22-byte zero prefix followed by a 26-byte hash suffix).
func (a Address) IsECDSA() bool {
	return hasZeroPrefixBytes(a[:], ECDSAZeroPrefixLength)
}

// IsLegacyCompatible reports whether the canonical 48-byte address can be
// represented losslessly using the historic 32-byte mainnet encoding.
//
// All pre-Aegis on-chain addresses were stored as either 20-byte aliases or
// 32-byte values. When those legacy encodings are canonicalized into the new
// 48-byte address type they are left-padded with zeros, which means every
// historical address has at least a 16-byte zero prefix in canonical form.
//
// This is intentionally broader than IsECDSA: mainnet already contains legacy
// 32-byte addresses whose payload is not an ECDSA-derived 26-byte suffix. Those
// addresses must continue to serialize back to 32 bytes in transactions, trie
// keys, logs and blooms or consensus hashes will change.
func (a Address) IsLegacyCompatible() bool {
	return hasZeroPrefixBytes(a[:], AddressLength-LegacyAddressLength)
}

// IsLegacyAirdropECDSA reports whether the canonical 48-byte address is the
// legacy zero-extended 20-byte ECDSA alias used by the airdrop machinery.
func (a Address) IsLegacyAirdropECDSA() bool {
	return a.IsECDSA() && hasZeroPrefixBytes(a[:], LegacyECDSAAirdropCanonicalZeroPrefixLength)
}

// CompactBytes returns the consensus/storage encoding of an address.
//
// Any address that fits losslessly into the historic 32-byte wire format must
// continue to use that encoding in order to preserve existing transaction
// hashes, trie keys and other mainnet consensus data. Only addresses that truly
// require the full 48-byte canonical form (e.g. MLDSA87/Q-addresses) are
// serialized as 48 bytes.
func (a Address) CompactBytes() []byte {
	if a.IsLegacyCompatible() {
		var compact [LegacyAddressLength]byte
		copy(compact[:], a[AddressLength-LegacyAddressLength:])
		return compact[:]
	}
	return a[:]
}

// EncodeRLP encodes addresses using the consensus/storage representation.
func (a Address) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, a.CompactBytes())
}

// DecodeRLP accepts legacy 20-byte aliases, legacy 32-byte compact ECDSA
// addresses and canonical 48-byte addresses.
func (a *Address) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Bytes()
	if err != nil {
		return err
	}
	if !isAcceptedAddressLength(len(raw)) {
		return fmt.Errorf("rlp: invalid address length %d", len(raw))
	}
	a.SetBytes(raw)
	return nil
}

func decodeAddressJSON(input []byte) ([]byte, error) {
	if !isString(input) {
		return nil, &json.UnmarshalTypeError{Value: "non-string", Type: addressT}
	}
	var s string
	if err := json.Unmarshal(input, &s); err != nil {
		return nil, err
	}
	return decodeAddressText([]byte(s), true)
}
