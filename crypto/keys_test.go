package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), PrivKeyLen)

	pubKey := privKey.Public()
	assert.Equal(t, len(pubKey.Bytes()), PubKeyLen)
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()

	msg := []byte("fuq yo couch!")
	sig := privKey.Sign(msg)

	assert.True(t, sig.Verify(pubKey, msg))

	// Test with invalid message
	assert.False(t, sig.Verify(pubKey, []byte("kiss yo couch!")))

	// Test with invalid public key
	altPrivKey := GeneratePrivateKey()
	altPubKey := altPrivKey.Public()

	assert.False(t, sig.Verify(altPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	addr := pubKey.Address()

	// fmt.Println("\n*** >>> [addr] -", addr)

	assert.Equal(t, AddressLen, len(addr.Bytes()))
}

func TestPrivateKeyFromString(t *testing.T) {

	// seed = make([]byte, 32)
	// io.ReadFull(rand.Reader, seed)
	// fmt.Println(hex.EncodeToString(seed))

	var (
		seed       = "d9822b1297a81035af59e88f40cc26d12d9ed77314d2c0ebac1b83f12d34d36c"
		addressStr = "156577acbd7ebc143352a1dcf4098db5d2fa1b31"
		privKey    = NewPrivateKeyFromString(seed)
	)

	address := privKey.Public().Address()

	// fmt.Println("\n*** >>> [add from string] -", address)

	assert.Equal(t, addressStr, address.String())
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))

}
