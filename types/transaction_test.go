package types

import (
	"testing"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {

	balance := 100
	sendAmt := 5

	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.PubKey().Address().Bytes()

	destPrivKey := crypto.GeneratePrivateKey()
	destAddress := destPrivKey.PubKey().Address().Bytes()

	input := &proto.TxInput{
		PrevOutIndex: 0,
		PrevTxHash:   util.RandomHash(),
		PubKey:       fromPrivKey.PubKey().Bytes(),
	}

	outputA := &proto.TxOutput{
		Amount:  uint64(sendAmt),
		Address: destAddress,
	}

	outputB := &proto.TxOutput{
		Amount:  uint64(balance - sendAmt),
		Address: fromAddress,
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{outputA, outputB},
	}

	sig := SignTransaction(fromPrivKey, tx)

	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(tx))
	// fmt.Printf("\n*** >>> [tx]\n%+v\n", tx)
}
