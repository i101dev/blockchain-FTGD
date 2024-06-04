package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/types"
)

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
type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer) *Chain {

	newChain := &Chain{
		blockStore: bs,
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
	return c.blockStore.Put(b)
}

func (c *Chain) AddBlock(b *proto.Block) error {

	if err := c.ValidateBlock(b); err != nil {
		return err
	}

	c.headers.Add(b.Header)

	return c.blockStore.Put(b)
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

	return nil
}

func createGenesisBlock() *proto.Block {

	privKey := crypto.GeneratePrivateKey()

	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	types.SignBlock(privKey, block)

	return block
}
