package types

import (
	"testing"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	assert.Equal(t, 32, len(hash))
}

func TestSignBlock(t *testing.T) {

	var (
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.Public()
		block   = util.RandomBlock()
	)

	sig := SignBlock(privKey, block)

	assert.Equal(t, crypto.SignatureLen, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))
}
