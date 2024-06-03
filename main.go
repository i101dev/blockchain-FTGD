package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/i101dev/blocker/node"
	"github.com/i101dev/blocker/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	originNode    = ":3000"
	startingPeers = []string{originNode}
)

func main() {

	node1 := makeNode(originNode, []string{})
	time.Sleep(time.Second * 2)

	node2 := makeNode(":4000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 4000 - %+v\n", node1.GetPeerList())

	node3 := makeNode(":5000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 5000 - %+v\n", node1.GetPeerList())

	node4 := makeNode(":6000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 6000 - %+v\n", node1.GetPeerList())

	node5 := makeNode(":7000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 7000 - %+v\n", node1.GetPeerList())

	node6 := makeNode(":8000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 8000 - %+v\n", node1.GetPeerList())

	node7 := makeNode(":9000", startingPeers)
	time.Sleep(time.Second * 2)
	// fmt.Printf("\nnode 1 peers - POST 8000 - %+v\n", node1.GetPeerList())

	fmt.Println("\n----------------------------------------------------------------------------")
	fmt.Printf("\nnode 1 peers - %+v\n", node1.GetPeerList())
	fmt.Printf("node 2 peers - %+v\n", node2.GetPeerList())
	fmt.Printf("node 3 peers - %+v\n", node3.GetPeerList())
	fmt.Printf("node 4 peers - %+v\n", node4.GetPeerList())
	fmt.Printf("node 5 peers - %+v\n", node5.GetPeerList())
	fmt.Printf("node 6 peers - %+v\n", node6.GetPeerList())
	fmt.Printf("node 7 peers - %+v\n", node7.GetPeerList())

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {

	n := node.NewNode(listenAddr)

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
	client, err := grpc.NewClient(originNode, opts...)

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
