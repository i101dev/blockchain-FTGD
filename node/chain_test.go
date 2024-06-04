package node

import (
	"encoding/hex"
	"fmt"
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

	block := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())

	require.Nil(t, err)

	block.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, block)

	return block
}

func TestNewChain(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore(), NewMemoryUTXOStore())
	assert.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	assert.Nil(t, err)
}

func TestChainHeight(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore(), NewMemoryUTXOStore())

	for i := 0; i < 100; i++ {

		b := RandomBlock(t, chain)

		require.Nil(t, chain.AddBlock(b))
		require.Equal(t, chain.Height(), i+1)
	}
}

func TestAddBlock(t *testing.T) {

	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore(), NewMemoryUTXOStore())

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

func TestAddBlockWithTX(t *testing.T) {

	var (
		receiverPubKey = crypto.GeneratePrivateKey().PubKey().Address().Bytes()
		senderPrivKey  = crypto.NewPrivateKeyFromSeedStr(originSeed)

		chain = NewChain(NewMemoryBlockStore(), NewMemoryTXStore(), NewMemoryUTXOStore())
		block = RandomBlock(t, chain)
	)

	ftt, err := chain.txStore.Get("9b35f571bcc6c3718df2ecae5c5d9ae0086f5256734f80455da6c0f147fe0201")
	assert.Nil(t, err)

	inputs := []*proto.TxInput{
		{
			PrevOutIndex: 0,
			PrevTxHash:   types.HashTransaction(ftt),
			PubKey:       senderPrivKey.PubKey().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:  100,
			Address: receiverPubKey,
		},
		{
			Amount:  23,
			Address: senderPrivKey.PubKey().Address().Bytes(),
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}

	sig := types.SignTransaction(senderPrivKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))

	txHash := hex.EncodeToString(types.HashTransaction(tx))

	fetchedTx, err := chain.txStore.Get(txHash)
	assert.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)

	// Check to see if there exists an unspent UTXO -------

	address := crypto.AddressFromBytes(tx.Outputs[1].Address)
	key := fmt.Sprintf("%s_%s", address, txHash)

	utxo, err := chain.utxoStore.Get(key)
	fmt.Println("\n*** >>> [UTXO] -", utxo)
	assert.Nil(t, err)
	// assert.Equal(t, outputs[1].Amount, utxo.Amount)
}
