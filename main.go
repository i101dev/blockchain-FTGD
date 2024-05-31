package main

import (
	"context"
	"fmt"
	"log"

	"github.com/i101dev/blocker/node"
	"github.com/i101dev/blocker/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port = ":3000"
)

func main() {

	startFini := make(chan struct{})

	node1 := makeNode(":3000", []string{}, startFini)
	<-startFini

	node2 := makeNode(":4000", []string{":3000"}, startFini)
	<-startFini

	node3 := makeNode(":5000", []string{":3000", ":4000"}, startFini)
	<-startFini

	fmt.Printf("\nnode 1 peers - %+v\n", node1.PeerList())
	fmt.Printf("node 2 peers - %+v\n", node2.PeerList())
	fmt.Printf("node 3 peers - %+v\n", node3.PeerList())

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string, startDone chan struct{}) *node.Node {

	n := node.NewNode(listenAddr)

	go func() {
		if err := n.Start(startDone); err != nil {
			log.Fatal("Failed to start node - ", err)
		}
	}()

	if len(bootstrapNodes) > 0 {
		if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
			log.Fatal("\n*** >>> [BootstrapNetwork] - ", err)
		}
	}

	return n
}

func makeTransaction() {

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	client, err := grpc.NewClient(port, opts...)

	if err != nil {
		log.Fatal("\n*** >>> [grpc.NewClient] - FAIL -", err)
	}

	c := proto.NewNodeClient(client)

	txn := &proto.Transaction{
		Version: 1,
	}

	_, err = c.HandleTX(context.TODO(), txn)

	if err != nil {
		log.Fatal("\n*** >>> [makeTransaction] - FAIL -", err)
	}
}

func makeHandshake() {

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	client, err := grpc.NewClient(port, opts...)

	if err != nil {
		log.Fatal("\n*** >>> [grpc.NewClient] - FAIL -", err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version:    "myChain-0.1",
		Height:     13,
		ListenAddr: ":4000",
	}

	_, err = c.Handshake(context.TODO(), version)

	if err != nil {
		log.Fatal("\n*** >>> [makeHandshake] - FAIL -", err)
	}
}
