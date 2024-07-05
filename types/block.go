package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cbergoon/merkletree"
	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"

	pb "google.golang.org/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {

	if len(block.Transactions) > 0 {

		tree, err := GetMerkleTree(block)

		if err != nil {
			panic(err)
		}

		block.Header.RootHash = tree.MerkleRoot()
	}

	blockHash := HashBlock(block)
	blockSig := pk.Sign(blockHash)

	block.PublicKey = pk.PubKey().Bytes()
	block.Signature = blockSig.Bytes()

	return blockSig
}

func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(h *proto.Header) []byte {

	b, err := pb.Marshal(h)

	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}

func VerifyBlock(b *proto.Block) bool {

	// if len(b.Transactions) > 0 {
	// if !VerifyRootHash(b) {
	// 	fmt.Println("\n*** >>> INVALID ROOT HASH <<< ***")
	// 	return false
	// }
	// }

	if len(b.PublicKey) != crypto.PubKeyLen {
		fmt.Println("\n*** >>> INVALID PUBLIC KEY LENGTH <<< ***")
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		fmt.Println("\n*** >>> INVALID SIGNATURE LENGTH <<< ***")
		return false
	}

	pubKey := crypto.PubKeyFromBytes(b.PublicKey)
	sig := crypto.SignatureFromBytes(b.Signature)
	hash := HashBlock(b)

	return sig.Verify(pubKey, hash)
}

// -------------------------------------------------------
type TxHash struct {
	hash []byte
}

func NewTxHash(h []byte) TxHash {
	return TxHash{hash: h}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

func VerifyRootHash(b *proto.Block) bool {

	merkleTree, err := GetMerkleTree(b)
	if err != nil {
		return false
	}

	treeIsValid, err := merkleTree.VerifyTree()

	if err != nil || !treeIsValid {
		return false
	}

	return bytes.Equal(b.Header.RootHash, merkleTree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {

	// if len(b.Transactions) == 0 {

	// }

	list := make([]merkletree.Content, len(b.Transactions))

	for i := 0; i < len(b.Transactions); i++ {
		txHash := HashTransaction(b.Transactions[i])
		list[i] = NewTxHash(txHash)
	}

	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
}
