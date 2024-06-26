package types

import (
	"testing"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	assert.Equal(t, 32, len(hash))
}

func TestSignVerifyBlock(t *testing.T) {

	var (
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.PubKey()
		block   = util.RandomBlock()
	)

	sig := SignBlock(privKey, block)

	assert.Equal(t, crypto.SignatureLen, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))

	assert.Equal(t, block.PublicKey, pubKey.Bytes())
	assert.Equal(t, block.Signature, sig.Bytes())

	assert.True(t, VerifyBlock(block))

	invalidPrivKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivKey.PubKey().Bytes()
	assert.False(t, VerifyBlock(block))
}

func TestCalculateRootHash(t *testing.T) {

	block := util.RandomBlock()
	privKey := crypto.GeneratePrivateKey()

	tx := &proto.Transaction{
		Version: 1,
	}

	block.Transactions = append(block.Transactions, tx)
	SignBlock(privKey, block)
	assert.True(t, VerifyRootHash(block))
	assert.Equal(t, 32, len(block.Header.RootHash))

	// _, err := GetMerkleTree(block)
	// assert.Nil(t, err)
	// fmt.Println("block -", len(block.Header.RootHash))
}
