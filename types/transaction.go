package types

import (
	"crypto/sha256"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"

	pb "google.golang.org/protobuf/proto"
)

func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {

	b, err := pb.Marshal(tx)

	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}

func VerifyTransaction(tx *proto.Transaction) bool {

	for _, input := range tx.Inputs {

		sig := crypto.SignatureFromBytes(input.Signature)
		pubKey := crypto.PubKeyFromBytes(input.PubKey)

		input.Signature = nil

		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
	}

	return true
}
