package node

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/i101dev/blocker/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Node struct {
	listenAddr string
	version    string

	peerLock sync.RWMutex
	peerList []string
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewNode(listenAddr string) *Node {

	return &Node{
		listenAddr: listenAddr,
		peerList:   make([]string, 0),
		peers:      make(map[proto.NodeClient]*proto.Version),
		version:    "myChain-0.1",
	}
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	client, err := grpc.NewClient(listenAddr, opts...)

	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(client), nil
}

func (n *Node) Start(startFini chan struct{}) error {

	opts := []grpc.ServerOption{}
	gRPCserver := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return err
	}

	proto.RegisterNodeServer(gRPCserver, n)
	// fmt.Println("*** >>> Server node active", n.listenAddr)

	startFini <- struct{}{} // Signal that this node has started

	return gRPCserver.Serve(ln)
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	fmt.Printf("\n(%s) - New peer: (%s) - height: (%d)", n.listenAddr, v.ListenAddr, v.Height)

	n.peers[c] = v
	n.peerList = append(n.peerList, v.ListenAddr)
}

func (n *Node) deletePeer(c proto.NodeClient, listenAddr string) {

	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)

	// Find the index of the listenAddr in peerList
	index := -1
	for i, addr := range n.peerList {
		if addr == listenAddr {
			index = i
			break
		}
	}

	// If the address is found, remove it from peerList
	if index != -1 {
		n.peerList = append(n.peerList[:index], n.peerList[index+1:]...)
	}
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {

	client, err := makeNodeClient(v.ListenAddr)

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Handshake @ %s, from %s", n.listenAddr, v.ListenAddr)

	n.addPeer(client, v)

	return n.getMetadata(), nil
}

func (n *Node) HandleTX(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("\n*** >>> received [tx] from:", peer)
	return &proto.Ack{}, nil
}

// --------------------------------------------------------------------------------------
func (n *Node) BootstrapNetwork(addrs []string) error {

	for _, addr := range addrs {

		c, err := makeNodeClient(addr)
		if err != nil {
			return err
		}

		version := n.getMetadata()
		// fmt.Printf("Shaking hand with version - %+v", version)

		v, err := c.Handshake(context.Background(), version)
		if err != nil {
			fmt.Println("\n*** >>> HANDSHAKE FAILED! - ", err)
			continue
		}

		n.addPeer(c, v)
	}

	return nil
}

func (n *Node) getMetadata() *proto.Version {
	return &proto.Version{
		ListenAddr: n.listenAddr,
		Version:    "blocker-0.1",
		Height:     0,
	}
}

func (n *Node) PeerList() []string {
	return n.peerList
}
