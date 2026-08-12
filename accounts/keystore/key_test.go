package keystore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/google/uuid"
	"github.com/ixios-io/ixiosSpark/accounts"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/crypto"
	"golang.org/x/crypto/sha3"
)

func mustGenerateMLDSA87Key(t *testing.T) (*mldsa87.PrivateKey, *mldsa87.PublicKey) {
	t.Helper()

	_, privateKey, err := mldsa87.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	publicKey, ok := privateKey.Public().(*mldsa87.PublicKey)
	if !ok {
		t.Fatalf("unexpected public key type %T", privateKey.Public())
	}
	return privateKey, publicKey
}

// marshalMLDSA87PlainKeyWithStoredAddress builds a plaintext keystore JSON blob
// whose "address" field is set to the caller-supplied hex. Used by the
// negative tests below to verify that non-canonical stored addresses are
// rejected on load.
func marshalMLDSA87PlainKeyWithStoredAddress(t *testing.T, privateKey *mldsa87.PrivateKey, storedAddressHex string) []byte {
	t.Helper()

	blob, err := json.Marshal(plainKeyJSON{
		Address:    storedAddressHex,
		PrivateKey: hex.EncodeToString(privateKey.Bytes()),
		Id:         uuid.New().String(),
		Version:    version,
		KeyType:    KeyTypeMLDSA87,
	})
	if err != nil {
		t.Fatalf("Marshal plain key JSON: %v", err)
	}
	return blob
}

// nonCanonicalMLDSA87Address computes the pre-canonicalisation "leading 48
// bytes of SHA3-512(pubkey)" address form. Kept only for use by the negative
// tests below; production code no longer knows how to produce this shape.
func nonCanonicalMLDSA87Address(pub []byte) common.Address {
	digest := sha3.Sum512(pub)
	var address common.Address
	copy(address[:], digest[:common.AddressLength])
	return address
}

func TestMLDSA87UnmarshalJSONAcceptsCanonicalStoredAddress(t *testing.T) {
	privateKey, publicKey := mustGenerateMLDSA87Key(t)
	canonicalAddress := crypto.MLDSA87PubkeyToAddress(publicKey.Bytes())

	blob := marshalMLDSA87PlainKeyWithStoredAddress(t, privateKey, hex.EncodeToString(canonicalAddress[:]))

	var key Key
	if err := json.Unmarshal(blob, &key); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if key.Address != canonicalAddress {
		t.Fatalf("unexpected address\nwant: %x\n got: %x", canonicalAddress, key.Address)
	}
}

func TestMLDSA87UnmarshalJSONRejectsNonCanonicalStoredAddress(t *testing.T) {
	privateKey, publicKey := mustGenerateMLDSA87Key(t)
	nonCanonical := nonCanonicalMLDSA87Address(publicKey.Bytes())

	blob := marshalMLDSA87PlainKeyWithStoredAddress(t, privateKey, hex.EncodeToString(nonCanonical[:]))

	var key Key
	err := json.Unmarshal(blob, &key)
	if err == nil {
		t.Fatal("UnmarshalJSON accepted a non-canonical stored address")
	}
	if !strings.Contains(err.Error(), "key content mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlainKeystoreGetKeyRejectsNonCanonicalMLDSA87Address(t *testing.T) {
	privateKey, publicKey := mustGenerateMLDSA87Key(t)
	nonCanonical := nonCanonicalMLDSA87Address(publicKey.Bytes())

	dir := t.TempDir()
	path := filepath.Join(dir, "key.json")
	blob := marshalMLDSA87PlainKeyWithStoredAddress(t, privateKey, hex.EncodeToString(nonCanonical[:]))
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	ks := keyStorePlain{keysDirPath: dir}
	if _, err := ks.GetKey(nonCanonical, path, ""); err == nil {
		t.Fatal("GetKey accepted a non-canonical stored address")
	}
}

func TestGetDecryptedKeyRejectsNonCanonicalMLDSA87Address(t *testing.T) {
	privateKey, publicKey := mustGenerateMLDSA87Key(t)
	nonCanonical := nonCanonicalMLDSA87Address(publicKey.Bytes())

	dir := t.TempDir()
	ks := NewPlaintextKeyStore(dir)

	path := filepath.Join(dir, "non-canonical-mldsa87.json")
	blob := marshalMLDSA87PlainKeyWithStoredAddress(t, privateKey, hex.EncodeToString(nonCanonical[:]))
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	account := accounts.Account{Address: nonCanonical, URL: accounts.URL{Scheme: KeyStoreScheme, Path: path}}
	ks.cache.add(account)

	if _, _, err := ks.getDecryptedKey(account, ""); err == nil {
		t.Fatal("getDecryptedKey accepted a non-canonical stored address")
	}
}
