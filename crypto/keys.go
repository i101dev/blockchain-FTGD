package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	SignatureLen = 64
	PrivKeyLen   = 64
	PubKeyLen    = 32
	SeedLen      = 32
	AddressLen   = 20
	Version      = 0x00
)

// -----------------------------------------------------------------
// -----------------------------------------------------------------
type PrivateKey struct {
	key ed25519.PrivateKey
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) PubKey() *PublicKey {
	b := make([]byte, PubKeyLen)
	copy(b, p.key[PubKeyLen:])
	return &PublicKey{
		key: b,
	}
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLen)

	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}

	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != SeedLen {
		panic("invalid seed length")
	}

	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromString(s string) *PrivateKey {

	b, err := hex.DecodeString(s)

	if err != nil {
		panic(err)
	}

	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeedStr(s string) *PrivateKey {
	seedBytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(seedBytes)
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------
type PublicKey struct {
	key ed25519.PublicKey
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

func (p *PublicKey) Address() Address {
	return Address{
		value: p.key[len(p.key)-AddressLen:],
	}
}

func PubKeyFromBytes(b []byte) *PublicKey {

	if len(b) != PubKeyLen {
		panic("invalid pubkey length")
	}

	return &PublicKey{
		key: ed25519.PublicKey(b),
	}
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------
type Signature struct {
	value []byte
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

func SignatureFromBytes(b []byte) *Signature {

	if len(b) != SignatureLen {
		panic("invalid signature length")
	}

	return &Signature{
		value: b,
	}
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------
type Address struct {
	value []byte
}

func (a Address) Bytes() []byte {
	return a.value
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

func AddressFromBytes(b []byte) Address {

	if len(b) != AddressLen {
		panic("invalid address length - not equal to 20")
	}

	return Address{
		value: b,
	}
}
