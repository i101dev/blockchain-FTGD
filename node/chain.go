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
const originSeed = "b72a9caf5a5c5e6b88ee6f25f053d07b43ddc263a034e2b8e7175e558c18a6ed"

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

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

func (list *HeaderList) Len() int {
	return len(list.headers)
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
		// fmt.Println("*** >>> New TX:", hex.EncodeToString(types.HashTransaction(tx)))
		if err := c.txStore.Put(tx); err != nil {
			return err
		}

		hash := hex.EncodeToString(types.HashTransaction(tx))

		// address_txHash
		for it, output := range tx.Outputs {

			utxo := &UTXO{
				Hash:     hash,
				Amount:   output.Amount,
				OutIndex: it,
				Spent:    false,
			}

			address := crypto.AddressFromBytes(output.Address)
			key := fmt.Sprintf("%s_%s", address, hash)

			if err := c.utxoStore.Put(key, utxo); err != nil {
				return err
			}
		}
	}

	return c.blockStore.Put(b)
}

func (c *Chain) AddBlock(b *proto.Block) error {

	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	return c.addBlock(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
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
		if !types.VerifyTransaction(tx) {
			return fmt.Errorf("invalid transaction signature")
		}

		// for _, input := range tx.Inputs {
		// }
	}

	return nil
}

func createGenesisBlock() *proto.Block {

	privKey := crypto.NewPrivateKeyFromSeedStr(originSeed)

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
