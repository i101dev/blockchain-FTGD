package node

import (
	"context"
	"fmt"

	"github.com/i101dev/blocker/proto"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version string
	height  int32
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	return &Node{
		version: "myChain-0.1",
	}
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {

	localVersion := &proto.Version{
		Version: n.version,
		Height:  n.height,
	}

	peer, _ := peer.FromContext(ctx)

	fmt.Printf("\n*** >>> received handshake from %s: %+v\n", v, peer.Addr)

	return localVersion, nil
}

func (n *Node) HandleTX(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("\n*** >>> received [tx] from:", peer)
	return &proto.Ack{}, nil
}
