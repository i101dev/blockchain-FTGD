package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/node"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	originNode    = ":3000"
	startingPeers = []string{originNode}
)

func main() {

	node1 := makeNode(originNode, []string{}, true)
	time.Sleep(time.Second * 2)

	node2 := makeNode(":4000", startingPeers, false)
	time.Sleep(time.Second * 2)

	node3 := makeNode(":5000", startingPeers, false)
	time.Sleep(time.Second * 2)

	// node4 := makeNode(":6000", startingPeers, false)
	// time.Sleep(time.Second * 2)

	// node5 := makeNode(":7000", startingPeers, false)
	// time.Sleep(time.Second * 2)

	// node6 := makeNode(":8000", startingPeers, false)
	// time.Sleep(time.Second * 2)

	// node7 := makeNode(":9000", startingPeers, false)
	// time.Sleep(time.Second * 2)

	fmt.Println("\n----------------------------------------------------------------------------")
	fmt.Printf("\nnode 1 peers - %+v\n", node1.GetPeerList())
	fmt.Printf("node 2 peers - %+v\n", node2.GetPeerList())
	fmt.Printf("node 3 peers - %+v\n", node3.GetPeerList())
	// fmt.Printf("node 4 peers - %+v\n", node4.GetPeerList())
	// fmt.Printf("node 5 peers - %+v\n", node5.GetPeerList())
	// fmt.Printf("node 6 peers - %+v\n", node6.GetPeerList())
	// fmt.Printf("node 7 peers - %+v\n", node7.GetPeerList())

	for {
		fmt.Println("\n----------------------------------------------------------------------------")
		fmt.Println("\n*** >>> Making transaction")
		time.Sleep(time.Millisecond * 400)
		makeTransaction()
	}

	// select {}
}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {

	cfg := node.ServerConfig{
		Version:    "blocker-0.1",
		ListenAddr: listenAddr,
		PrivateKey: nil,
	}

	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}

	n := node.NewNode(cfg)

	go func() {
		if err := n.Start(bootstrapNodes); err != nil {
			log.Fatal("Failed to start node - ", err)
		}
	}()

	return n
}

func makeTransaction() {

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	client, err := grpc.NewClient(originNode, opts...)

	if err != nil {
		log.Fatal("\n*** >>> [grpc.NewClient] - FAIL -", err)
	}

	c := proto.NewNodeClient(client)

	privKey := crypto.GeneratePrivateKey()

	txn := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PubKey:       privKey.PubKey().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.PubKey().Address().Bytes(),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err = c.HandleTX(ctx, txn)

	if err != nil {
		log.Fatal("\n*** >>> [makeTransaction] - FAIL -", err)
	}
}

// func makeHandshake() {

// 	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
// 	client, err := grpc.NewClient(originNode, opts...)

// 	if err != nil {
// 		log.Fatal("\n*** >>> [grpc.NewClient] - FAIL -", err)
// 	}

// 	c := proto.NewNodeClient(client)

// 	version := &proto.Version{
// 		Version:    "myChain-0.1",
// 		Height:     13,
// 		ListenAddr: ":4000",
// 	}

// 	_, err = c.Handshake(context.TODO(), version)

// 	if err != nil {
// 		log.Fatal("\n*** >>> [makeHandshake] - FAIL -", err)
// 	}
// }
