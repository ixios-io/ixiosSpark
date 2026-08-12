package crypto

import (
	"testing"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/ixios-io/ixiosSpark/common"
	"golang.org/x/crypto/sha3"
)

func TestMLDSA87PubkeyToAddressUsesSHA3512(t *testing.T) {
	pub := make([]byte, mldsa87.PublicKeySize)
	for i := range pub {
		pub[i] = byte((i*29 + 7) % 256)
	}

	sum := sha3.Sum512(pub)

	var want common.Address
	copy(want[:], sum[len(sum)-common.AddressLength:])

	got := MLDSA87PubkeyToAddress(pub)
	if got != want {
		t.Fatalf("unexpected MLDSA87 address\nwant: %x\n got: %x", want, got)
	}
	if got != PubkeyBytesToAddressWithType(0x01, pub) {
		t.Fatalf("generic MLDSA87 address derivation diverged from MLDSA87PubkeyToAddress")
	}

	legacyKeccak := Keccak512(pub)
	var legacy common.Address
	copy(legacy[:], legacyKeccak[:common.AddressLength])
	if got == legacy {
		t.Fatalf("MLDSA87 address derivation regressed to Keccak-512")
	}

	var first48 common.Address
	copy(first48[:], sum[:common.AddressLength])
	if got == first48 {
		t.Fatalf("MLDSA87 address derivation must use the trailing 48 bytes of SHA3-512 to match IxiosOrbit, not the leading 48 bytes")
	}

	rightTrimmed := common.BytesToAddress(sum[:])
	if got != rightTrimmed {
		t.Fatalf("MLDSA87 address derivation must match the trailing 48 bytes of SHA3-512")
	}
}

func TestECDSAPubkeyBytesToAddressKeepsLegacyLayout(t *testing.T) {
	pub := make([]byte, 64)
	for i := range pub {
		pub[i] = byte((i*17 + 3) % 256)
	}

	hash := Keccak256(pub)
	var want common.Address
	copy(want[common.ECDSAZeroPrefixLength:], hash[len(hash)-common.ECDSAAddressHashLength:])

	got := PubkeyBytesToAddressWithType(0x00, pub)
	if got != want {
		t.Fatalf("unexpected ECDSA address\nwant: %x\n got: %x", want, got)
	}
	if !got.IsECDSA() {
		t.Fatalf("ECDSA address lost its canonical zero-prefix layout")
	}
	if !got.IsLegacyCompatible() {
		t.Fatalf("ECDSA address should remain legacy-compatible")
	}
}
