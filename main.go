package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/i101dev/blocker/node"
	"github.com/i101dev/blocker/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port = ":3000"
)

func main() {

	ln, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatal(err)
	}

	node := node.NewNode()
	opts := []grpc.ServerOption{}
	gRPCserver := grpc.NewServer(opts...)

	proto.RegisterNodeServer(gRPCserver, node)
	fmt.Println("\n*** >>> Server node alive on port", port)

	go func() {
		for {
			time.Sleep(2 * time.Second)
			// makeTransaction()
			makeHandshake()
		}
	}()

	gRPCserver.Serve(ln)
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
		Version: "myChain-0.1",
		Height:  13,
	}

	_, err = c.Handshake(context.TODO(), version)

	if err != nil {
		log.Fatal("\n*** >>> [makeHandshake] - FAIL -", err)
	}
}
