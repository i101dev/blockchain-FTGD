package types

import (
	"crypto/sha256"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"

	pb "google.golang.org/protobuf/proto"
)

func VerifyBlock(b *proto.Block) bool {

	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}

	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PubKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {

	blockHash := HashBlock(block)
	blockSig := pk.Sign(blockHash)

	block.PublicKey = pk.PubKey().Bytes()
	block.Signature = blockSig.Bytes()

	return blockSig
}

// Returns SHA256 of the header
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
