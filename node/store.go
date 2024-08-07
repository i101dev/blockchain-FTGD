package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/types"
)

// ------------------------------------------------------------------------
type UTXOStorer interface {
	Put(*UTXO) error
	Get(string) (*UTXO, error)
}

type MemoryUTXOStore struct {
	lock sync.RWMutex
	data map[string]*UTXO
}

func NewMemoryUTXOStore() *MemoryUTXOStore {
	return &MemoryUTXOStore{
		data: make(map[string]*UTXO),
	}
}

func (s *MemoryUTXOStore) Get(hash string) (*UTXO, error) {

	s.lock.RLock()
	defer s.lock.RUnlock()

	utxo, ok := s.data[hash]
	if !ok {
		return nil, fmt.Errorf("failed to fetch UTXO with hash - %s", hash)
	}

	return utxo, nil
}

func (s *MemoryUTXOStore) Put(utxo *UTXO) error {

	s.lock.Lock()
	defer s.lock.Unlock()

	key := fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)

	s.data[key] = utxo

	return nil
}

// ------------------------------------------------------------------------

type TXStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}

type MemoryTXStore struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMemoryTXStore() *MemoryTXStore {
	return &MemoryTXStore{
		txx: make(map[string]*proto.Transaction),
	}
}

func (s *MemoryTXStore) Get(hash string) (*proto.Transaction, error) {

	s.lock.RLock()
	defer s.lock.RUnlock()

	tx, ok := s.txx[hash]

	if !ok {
		fmt.Printf("failed to get has TX by hash: %s", hash)
		return nil, fmt.Errorf("failed to get has TX by hash: %s", hash)
	}

	return tx, nil
}

func (s *MemoryTXStore) Put(tx *proto.Transaction) error {

	s.lock.Lock()
	defer s.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	s.txx[hash] = tx

	return nil
}

// ------------------------------------------------------------------------

type BlockStorer interface {
	PutBlock(*proto.Block) error
	GetBlock(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*proto.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}

func (s *MemoryBlockStore) PutBlock(b *proto.Block) error {

	s.lock.Lock()
	defer s.lock.Unlock()

	hash := hex.EncodeToString(types.HashBlock(b))
	s.blocks[hash] = b

	return nil
}

func (s *MemoryBlockStore) GetBlock(hash string) (*proto.Block, error) {

	s.lock.RLock()
	defer s.lock.RUnlock()

	block, ok := s.blocks[hash]

	if !ok {
		return nil, fmt.Errorf("failed to get block [%s]", hash)
	}

	return block, nil
}
