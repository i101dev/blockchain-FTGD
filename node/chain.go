package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/types"
)

// ----------------------------------------------------------------------------------
const originSeed string = "b72a9caf5a5c5e6b88ee6f25f053d07b43ddc263a034e2b8e7175e558c18a6ed"

// ----------------------------------------------------------------------------------
type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		panic("index too high")
	}
	return list.headers[index]
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

// ----------------------------------------------------------------
type UTXO struct {
	Hash     string
	OutIndex int
	Amount   uint64
	Spent    bool
}

// ----------------------------------------------------------------
type Chain struct {
	blockStore BlockStorer
	utxoStore  UTXOStorer
	txStore    TXStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer, ts TXStorer, us UTXOStorer) *Chain {

	newChain := &Chain{
		blockStore: bs,
		utxoStore:  us,
		txStore:    ts,
		headers:    NewHeaderList(),
	}

	newChain.addBlock(createGenesisBlock())

	return newChain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) addBlock(b *proto.Block) error {

	c.headers.Add(b.Header)

	for _, tx := range b.Transactions {

		if err := c.txStore.Put(tx); err != nil {
			return err
		}

		hash := hex.EncodeToString(types.HashTransaction(tx))

		for index, output := range tx.Outputs {

			utxo := &UTXO{
				Hash:     hash,
				Amount:   output.Amount,
				OutIndex: index,
				Spent:    false,
			}

			// key := fmt.Sprintf("%s_%d", hash, it)

			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}

		for _, input := range tx.Inputs {

			key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PrevTxHash), input.PrevOutIndex)
			utxo, err := c.utxoStore.Get(key)

			if err != nil {
				panic(err)
			}

			utxo.Spent = true

			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}

			// fmt.Println("\n-----------------------------------------------")
			// fmt.Printf("\n*** >>> [utxo.Hash] - %+v", utxo.Hash)
			// fmt.Printf("\n*** >>> [utxo.OutIndex] - %+v", utxo.OutIndex)
			// fmt.Printf("\n*** >>> [utxo.Amount] - %+v", utxo.Amount)
			// fmt.Printf("\n*** >>> [utxo.Spent] - %+v", utxo.Spent)
			// fmt.Println("\n-----------------------------------------------")
		}
	}

	return c.blockStore.PutBlock(b)
}

func (c *Chain) AddBlock(b *proto.Block) error {

	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	return c.addBlock(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.GetBlock(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {

	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - current height: (%d)", height, c.Height())
	}

	header := c.headers.Get(height)
	hash := types.HashHeader(header)

	return c.GetBlockByHash(hash)
}

func (c *Chain) ValidateBlock(newBlock *proto.Block) error {

	// Validate [newBlock] signature
	if !types.VerifyBlock(newBlock) {
		return fmt.Errorf("failed to verify block signature")
	}

	// Validate if the [prevHash] is the hash of the current block
	cBlock, err := c.GetBlockByHeight(c.Height())

	if err != nil {
		return fmt.Errorf("failed to get block by height: %+v", err)
	}

	cBlockHash := types.HashBlock(cBlock)

	if !bytes.Equal(cBlockHash, newBlock.Header.PrevHash) {
		return fmt.Errorf("previous block hash invalid")
	}

	for _, tx := range newBlock.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {

	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid transaction signature")
	}

	// Check if all inputs are unspent ----------------------------------
	hash := hex.EncodeToString(types.HashTransaction(tx))
	nInputs := len(tx.Inputs)

	sumInputs := 0
	for i := 0; i < nInputs; i++ {

		prevHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
		key := fmt.Sprintf("%s_%d", prevHash, i)

		utxo, err := c.utxoStore.Get(key)
		sumInputs += int(utxo.Amount)

		if err != nil {
			return err
		}

		if utxo.Spent {
			return fmt.Errorf("output [%d] of tx [%s] is spent", i, hash)
		}
	}

	sumOutputs := 0
	for _, output := range tx.Outputs {
		sumOutputs += int(output.Amount)
	}

	if sumInputs < sumOutputs {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}

func createGenesisBlock() *proto.Block {

	privKey := crypto.NewPrivateKeyFromString(originSeed)

	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	genesisTX := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  123,
				Address: privKey.PubKey().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, genesisTX)

	types.SignBlock(privKey, block)

	return block
}
