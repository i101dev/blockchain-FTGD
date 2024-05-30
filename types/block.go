package types

import (
	"crypto/sha256"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"

	pb "google.golang.org/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	blockHash := HashBlock(block)
	blockSig := pk.Sign(blockHash)
	return blockSig
}

// Returns SHA256 of the header
func HashBlock(block *proto.Block) []byte {

	b, err := pb.Marshal(block)

	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}
