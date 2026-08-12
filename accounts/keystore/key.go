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

package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa87"
	"github.com/google/uuid"
	"github.com/ixios-io/ixiosSpark/accounts"
	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/crypto"
	"golang.org/x/crypto/pbkdf2"
)

const (
	version = 3

	KeyTypeECDSA   = "ecdsa"
	KeyTypeMLDSA87 = "mldsa87"
)

type Key struct {
	Id uuid.UUID // Version 4 "random" for unique id not derived from key data
	// to simplify lookups we also store the address
	Address common.Address
	// key type determines which signing material is present.
	KeyType string
	// we only store privkey as pubkey/address can be derived from it.
	// key material in this struct is always plaintext while unlocked.
	PrivateKey        *ecdsa.PrivateKey
	MLDSA87PrivateKey []byte
}

type keyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr common.Address, filename string, auth string) (*Key, error)
	// Writes and encrypts the key.
	StoreKey(filename string, k *Key, auth string) error
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type plainKeyJSON struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privatekey"`
	Id         string `json:"id"`
	Version    int    `json:"version"`
	KeyType    string `json:"keytype,omitempty"`
}

type encryptedKeyJSONV3 struct {
	Address string     `json:"address"`
	Crypto  CryptoJSON `json:"crypto"`
	Id      string     `json:"id"`
	Version int        `json:"version"`
	KeyType string     `json:"keytype,omitempty"`
}

type encryptedKeyJSONV1 struct {
	Address string     `json:"address"`
	Crypto  CryptoJSON `json:"crypto"`
	Id      string     `json:"id"`
	Version string     `json:"version"`
	KeyType string     `json:"keytype,omitempty"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

// KDFParams holds both scrypt (N, R, P) and pbkdf2 (C, Prf) parameters.
// Fields that aren't used for a particular KDF can remain at zero values.
type KDFParams struct {
	// For scrypt
	N int `json:"n,omitempty"`
	R int `json:"r,omitempty"`
	P int `json:"p,omitempty"`

	// For pbkdf2
	C   int    `json:"c,omitempty"`
	Prf string `json:"prf,omitempty"`

	// For both
	DkLen int    `json:"dklen"`
	Salt  string `json:"salt"`
}

type CryptoJSON struct {
	Cipher       string           `json:"cipher"`
	CipherText   string           `json:"ciphertext"`
	CipherParams cipherparamsJSON `json:"cipherparams"`
	KDF          string           `json:"kdf"`
	KDFParams    KDFParams        `json:"kdfparams"`
	MAC          string           `json:"mac"`
}

func normalizeKeyType(keyType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(keyType)) {
	case "", KeyTypeECDSA:
		return KeyTypeECDSA, nil
	case KeyTypeMLDSA87, "ml-dsa87", "ml_dsa87", "mldsa-87", "ml-dsa-87":
		return KeyTypeMLDSA87, nil
	default:
		return "", fmt.Errorf("unsupported key type %q", keyType)
	}
}

func (k *Key) resolvedKeyType() string {
	if keyType, err := normalizeKeyType(k.KeyType); err == nil {
		return keyType
	}
	if len(k.MLDSA87PrivateKey) != 0 {
		return KeyTypeMLDSA87
	}
	return KeyTypeECDSA
}

func (k *Key) isECDSA() bool {
	return k.resolvedKeyType() == KeyTypeECDSA
}

func (k *Key) isMLDSA87() bool {
	return k.resolvedKeyType() == KeyTypeMLDSA87
}

func (k *Key) privateKeyBytes() ([]byte, error) {
	switch k.resolvedKeyType() {
	case KeyTypeECDSA:
		if k.PrivateKey == nil {
			return nil, errors.New("missing ECDSA private key")
		}
		return crypto.FromECDSA(k.PrivateKey), nil
	case KeyTypeMLDSA87:
		if len(k.MLDSA87PrivateKey) == 0 {
			return nil, errors.New("missing MLDSA87 private key")
		}
		return common.CopyBytes(k.MLDSA87PrivateKey), nil
	default:
		return nil, fmt.Errorf("unsupported key type %q", k.KeyType)
	}
}

func (k *Key) mldsa87PrivateKey() (*mldsa87.PrivateKey, error) {
	if !k.isMLDSA87() {
		return nil, fmt.Errorf("key type %q is not MLDSA87", k.resolvedKeyType())
	}
	var privateKey mldsa87.PrivateKey
	if err := privateKey.UnmarshalBinary(k.MLDSA87PrivateKey); err != nil {
		return nil, err
	}
	return &privateKey, nil
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	keyBytes, err := k.privateKeyBytes()
	if err != nil {
		return nil, err
	}
	jStruct := plainKeyJSON{
		Address:    hex.EncodeToString(k.Address[:]),
		PrivateKey: hex.EncodeToString(keyBytes),
		Id:         k.Id.String(),
		Version:    version,
		KeyType:    k.resolvedKeyType(),
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) UnmarshalJSON(data []byte) error {
	var keyJSON plainKeyJSON
	if err := json.Unmarshal(data, &keyJSON); err != nil {
		return err
	}

	u, err := uuid.Parse(keyJSON.Id)
	if err != nil {
		return err
	}
	k.Id = u

	keyType, err := normalizeKeyType(keyJSON.KeyType)
	if err != nil {
		return err
	}
	k.KeyType = keyType

	addr, err := hex.DecodeString(keyJSON.Address)
	if err != nil {
		return err
	}
	k.Address = common.BytesToAddress(addr)

	switch keyType {
	case KeyTypeECDSA:
		privkey, err := crypto.HexToECDSA(keyJSON.PrivateKey)
		if err != nil {
			return err
		}
		derivedAddress := crypto.PubkeyToAddress(privkey.PublicKey)
		if k.Address != (common.Address{}) && k.Address != derivedAddress {
			return fmt.Errorf("key content mismatch: have account %x, want %x", derivedAddress, k.Address)
		}
		k.Address = derivedAddress
		k.PrivateKey = privkey
		k.MLDSA87PrivateKey = nil
	case KeyTypeMLDSA87:
		privateKeyBytes, err := hex.DecodeString(keyJSON.PrivateKey)
		if err != nil {
			return err
		}
		var privateKey mldsa87.PrivateKey
		if err := privateKey.UnmarshalBinary(privateKeyBytes); err != nil {
			return err
		}
		publicKey, ok := privateKey.Public().(*mldsa87.PublicKey)
		if !ok {
			return errors.New("invalid MLDSA87 public key type")
		}
		derivedAddress := crypto.MLDSA87PubkeyToAddress(publicKey.Bytes())
		if k.Address != (common.Address{}) && k.Address != derivedAddress {
			return fmt.Errorf("key content mismatch: have account %x, want %x", derivedAddress, k.Address)
		}
		k.Address = derivedAddress
		k.PrivateKey = nil
		k.MLDSA87PrivateKey = common.CopyBytes(privateKeyBytes)
	default:
		return fmt.Errorf("unsupported key type %q", keyType)
	}

	return nil
}

func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *Key {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("Could not create random uuid: %v", err))
	}
	key := &Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		KeyType:    KeyTypeECDSA,
		PrivateKey: privateKeyECDSA,
	}
	return key
}

func newKeyFromMLDSA87(privateKeyMLDSA87 *mldsa87.PrivateKey) *Key {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("Could not create random uuid: %v", err))
	}
	publicKey, ok := privateKeyMLDSA87.Public().(*mldsa87.PublicKey)
	if !ok {
		panic("unexpected MLDSA87 public key type")
	}
	return &Key{
		Id:                id,
		Address:           crypto.MLDSA87PubkeyToAddress(publicKey.Bytes()),
		KeyType:           KeyTypeMLDSA87,
		MLDSA87PrivateKey: privateKeyMLDSA87.Bytes(),
	}
}

// NewKeyForDirectICAP generates a key whose address fits into < 155 bits so it can fit
// into the Direct ICAP spec. for simplicity and easier compatibility with other libs, we
// retry until the first byte is 0.
func NewKeyForDirectICAP(rand io.Reader) *Key {
	randBytes := make([]byte, 64)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic("key generation: could not read from random source: " + err.Error())
	}
	reader := bytes.NewReader(randBytes)
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), reader)
	if err != nil {
		panic("key generation: ecdsa.GenerateKey failed: " + err.Error())
	}
	key := newKeyFromECDSA(privateKeyECDSA)
	if !strings.HasPrefix(key.Address.Hex(), "0x00") {
		return NewKeyForDirectICAP(rand)
	}
	return key
}

func newKey(rand io.Reader) (*Key, error) {
	return newKeyWithType(rand, KeyTypeECDSA)
}

func newKeyWithType(rand io.Reader, keyType string) (*Key, error) {
	normalizedKeyType, err := normalizeKeyType(keyType)
	if err != nil {
		return nil, err
	}
	switch normalizedKeyType {
	case KeyTypeECDSA:
		privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
		if err != nil {
			return nil, err
		}
		return newKeyFromECDSA(privateKeyECDSA), nil
	case KeyTypeMLDSA87:
		_, privateKeyMLDSA87, err := mldsa87.GenerateKey(rand)
		if err != nil {
			return nil, err
		}
		return newKeyFromMLDSA87(privateKeyMLDSA87), nil
	default:
		return nil, fmt.Errorf("unsupported key type %q", keyType)
	}
}

func storeNewKey(ks keyStore, rand io.Reader, auth string) (*Key, accounts.Account, error) {
	return storeNewKeyWithType(ks, rand, auth, KeyTypeECDSA)
}

func storeNewKeyWithType(ks keyStore, rand io.Reader, auth string, keyType string) (*Key, accounts.Account, error) {
	key, err := newKeyWithType(rand, keyType)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	a := accounts.Account{
		Address: key.Address,
		URL:     accounts.URL{Scheme: KeyStoreScheme, Path: ks.JoinPath(keyFileName(key.Address))},
	}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		zeroKeyMaterial(key)
		return nil, a, err
	}
	return key, a, err
}

func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return "", err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := os.CreateTemp(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func writeKeyFile(file string, content []byte) error {
	name, err := writeTemporaryKeyFile(file, content)
	if err != nil {
		return err
	}
	return os.Rename(name, file)
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileName(keyAddr common.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

// creates a Key and stores that in the given KeyStore by decrypting a presale key JSON
func importKey(keyStore keyStore, keyJSON []byte, password string) (accounts.Account, *Key, error) {
	key, err := decryptKey(keyJSON, password)
	if err != nil {
		return accounts.Account{}, nil, err
	}
	key.Id, err = uuid.NewRandom()
	if err != nil {
		return accounts.Account{}, nil, err
	}
	a := accounts.Account{
		Address: key.Address,
		URL: accounts.URL{
			Scheme: KeyStoreScheme,
			Path:   keyStore.JoinPath(keyFileName(key.Address)),
		},
	}
	err = keyStore.StoreKey(a.URL.Path, key, password)
	return a, key, err
}

func decryptKey(fileContent []byte, password string) (key *Key, err error) {
	preSaleKeyStruct := struct {
		EncSeed string
		EthAddr string
		Email   string
		BtcAddr string
	}{}
	err = json.Unmarshal(fileContent, &preSaleKeyStruct)
	if err != nil {
		return nil, err
	}
	encSeedBytes, err := hex.DecodeString(preSaleKeyStruct.EncSeed)
	if err != nil {
		return nil, errors.New("invalid hex in encSeed")
	}
	if len(encSeedBytes) < 16 {
		return nil, errors.New("invalid encSeed, too short")
	}
	iv := encSeedBytes[:16]
	cipherText := encSeedBytes[16:]
	/*
		See https://github.com/ethereum/pyethsaletool

		pyethsaletool generates the encryption key from password by
		2000 rounds of PBKDF2 with HMAC-SHA-256 using password as salt (:().
		16 byte key length within PBKDF2 and resulting key is used as AES key
	*/
	passBytes := []byte(password)
	derivedKey := pbkdf2.Key(passBytes, passBytes, 2000, 16, sha256.New)
	plainText, err := aesCBCDecrypt(derivedKey, cipherText, iv)
	if err != nil {
		return nil, err
	}
	ethPriv := crypto.Keccak256(plainText)
	ecKey := crypto.ToECDSAUnsafe(ethPriv)

	key = &Key{
		Id:         uuid.UUID{},
		Address:    crypto.PubkeyToAddress(ecKey.PublicKey),
		KeyType:    KeyTypeECDSA,
		PrivateKey: ecKey,
	}
	derivedAddr := hex.EncodeToString(key.Address[common.AddressLength-common.LegacyECDSAAirdropAddressHashLength:])
	expectedAddr := preSaleKeyStruct.EthAddr
	if derivedAddr != expectedAddr {
		err = fmt.Errorf("decrypted addr '%s' not equal to expected addr '%s'", derivedAddr, expectedAddr)
	}
	return key, err
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func aesCBCDecrypt(key, cipherText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	paddedPlaintext := make([]byte, len(cipherText))
	decrypter.CryptBlocks(paddedPlaintext, cipherText)
	plaintext := pkcs7Unpad(paddedPlaintext)
	if plaintext == nil {
		return nil, ErrDecrypt
	}
	return plaintext, err
}

// From https://leanpub.com/gocrypto/read#leanpub-auto-block-cipher-modes
func pkcs7Unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}
