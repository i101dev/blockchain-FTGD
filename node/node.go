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
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func NewNode(listenAddr string) *Node {

	return &Node{
		listenAddr: listenAddr,
		// peerList:   make([]string, 0),
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "myChain-0.1",
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

func (n *Node) Start(bootstrapNodes []string) error {

	opts := []grpc.ServerOption{}
	gRPCserver := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return err
	}

	proto.RegisterNodeServer(gRPCserver, n)
	// fmt.Println("*** >>> Server node active", n.listenAddr)

	// bootstrap the network with a list of curated nodes
	if len(bootstrapNodes) > 0 {
		// if err := n.bootstrapNetwork(bootstrapNodes); err != nil {
		// 	log.Fatal("\n*** >>> [BootstrapNetwork] - ", err)
		// }
		go n.bootstrapNetwork(bootstrapNodes)
	}

	return gRPCserver.Serve(ln)
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// Handler logic where it is decided whether or reject the incoming node connection

	n.peers[c] = v

	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	fmt.Printf("\n(%s) - New peer: (%s) - height: (%d)", n.listenAddr, v.ListenAddr, v.Height)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {

	client, err := makeNodeClient(v.ListenAddr)

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Handshake @ %s, from %s", n.listenAddr, v.ListenAddr)

	n.addPeer(client, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTX(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("\n*** >>> received [tx] from:", peer)
	return &proto.Ack{}, nil
}

// --------------------------------------------------------------------------------------
func (n *Node) bootstrapNetwork(addrs []string) error {

	for _, addr := range addrs {

		if !n.canConnectWith(addr) {
			continue
		}

		fmt.Printf("\ndialing remote node - local: %s - remote: %s", n.listenAddr, addr)

		c, v, err := n.dialRemoteNode(addr)

		if err != nil {
			return err
		}

		n.addPeer(c, v)
	}

	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {

	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		fmt.Println("\n*** >>> HANDSHAKE FAILED! - ", err)
		return nil, nil, err
	}

	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		ListenAddr: n.listenAddr,
		Version:    "blocker-0.1",
		Height:     0,
		PeerList:   n.GetPeerList(),
	}
}

func (n *Node) GetPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	peerList := []string{}
	for _, version := range n.peers {
		peerList = append(peerList, version.ListenAddr)
	}

	return peerList
}

func (n *Node) canConnectWith(addr string) bool {

	if n.listenAddr == addr {
		return false
	}

	connectedPeers := n.GetPeerList()

	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}

	return true
}
