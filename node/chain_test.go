package node

import (
	"testing"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/types"
	"github.com/i101dev/blocker/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RandomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	b := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	b.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, b)
	return b
}

func TestNewChain(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore())
	assert.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	assert.Nil(t, err)
}

func TestChainHeight(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {

		b := RandomBlock(t, chain)

		require.Nil(t, chain.AddBlock(b))
		require.Equal(t, chain.Height(), i+1)
	}

}

func TestAddBlock(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore())

	for i := 0; i < 100; i++ {

		block := RandomBlock(t, chain)

		blockHash := types.HashBlock(block)
		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}
