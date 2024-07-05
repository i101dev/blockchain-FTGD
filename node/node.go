package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/i101dev/blocker/crypto"
	"github.com/i101dev/blocker/proto"
	"github.com/i101dev/blocker/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

// --------------------------------------------------------------
const blockTime = time.Second * 5

// --------------------------------------------------------------

type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (pool *Mempool) Clear() []*proto.Transaction {

	pool.lock.Lock()
	defer pool.lock.Unlock()

	txx := make([]*proto.Transaction, len(pool.txx))

	it := 0
	for k, v := range pool.txx {
		delete(pool.txx, k)
		txx[it] = v
		it++
	}

	return txx
}

func (pool *Mempool) Len() int {

	pool.lock.RLock()
	defer pool.lock.Unlock()

	return len(pool.txx)
}

func (pool *Mempool) Has(tx *proto.Transaction) bool {

	pool.lock.RLock()
	defer pool.lock.RUnlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := pool.txx[hash]
	return ok
}

func (pool *Mempool) Add(tx *proto.Transaction) bool {

	if pool.Has(tx) {
		return false
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	pool.txx[hash] = tx

	return true
}

// ----------------------------------------------------------------------

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	// listenAddr string
	// version    string
	ServerConfig

	peerLock sync.RWMutex
	peerList map[proto.NodeClient]*proto.Version

	mempool *Mempool

	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {

	return &Node{
		peerList:     make(map[proto.NodeClient]*proto.Version),
		mempool:      NewMempool(),
		ServerConfig: cfg,
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

	ln, err := net.Listen("tcp", n.ListenAddr)
	if err != nil {
		return err
	}

	proto.RegisterNodeServer(gRPCserver, n)

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return gRPCserver.Serve(ln)
}

func (n *Node) addPeer(client proto.NodeClient, nodeDat *proto.Version) {

	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	for _, peerVersion := range n.peerList {
		if peerVersion.ListenAddr == nodeDat.ListenAddr {
			return
		}
	}

	n.peerList[client] = nodeDat

	if len(nodeDat.PeerList) > 0 {
		go n.bootstrapNetwork(nodeDat.PeerList)
	}

	fmt.Printf("\n(%s) - New peer: (%s) - height: (%d)", n.ListenAddr, nodeDat.ListenAddr, nodeDat.Height)
}

// func (n *Node) deletePeer(c proto.NodeClient) {
// 	n.peerLock.Lock()
// 	defer n.peerLock.Unlock()
// 	delete(n.peerList, c)
// }

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {

	client, err := makeNodeClient(v.ListenAddr)

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Handshake @ %s, from %s", n.ListenAddr, v.ListenAddr)

	n.addPeer(client, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTX(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {

	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {

		fmt.Printf("\n*** >>> (%s) received [tx] from peer address: (%s)", n.ListenAddr, peer.Addr)
		fmt.Printf("\n*** >>> [hash] - %s", hash)

		go func() {
			if err := n.broadcast(tx); err != nil {
				log.Fatal("\n*** >>> BROADCAST ERROR <<< ***", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) broadcast(msg any) error {

	for peer := range n.peerList {

		switch v := msg.(type) {

		case *proto.Transaction:
			_, err := peer.HandleTX(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) validatorLoop() {

	fmt.Print("\n**** >>> Starting Validator Loop <<< ***")

	ticker := time.NewTicker(blockTime)

	for {
		<-ticker.C

		txx := n.mempool.Clear()

		fmt.Printf("\n*** >>> CREATE NEW BLOCK <<< *** || lenTx: (%d)", len(txx))
	}
}

// --------------------------------------------------------------------------------------

func (n *Node) bootstrapNetwork(knownAddres []string) error {

	var wg sync.WaitGroup

	for _, addr := range knownAddres {

		if !n.canConnectWith(addr) {
			continue
		}

		wg.Add(1)

		go func(a string) {

			// fmt.Printf("\ndialing remote node - local: %s - remote: %s", n.ListenAddr, addr)

			defer wg.Done()

			c, v, err := n.dialRemoteNode(a)

			if err != nil {
				log.Printf("\nFailed to dial remote node: %s - %v", a, err)
				return
			}

			n.addPeer(c, v)

		}(addr)
	}

	wg.Wait()

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
		ListenAddr: n.ListenAddr,
		Version:    "blocker-0.1",
		Height:     0,
		PeerList:   n.GetPeerList(),
	}
}

func (n *Node) GetPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	peerList := []string{}
	for _, version := range n.peerList {
		peerList = append(peerList, version.ListenAddr)
	}

	return peerList
}

func (n *Node) canConnectWith(addr string) bool {

	if n.ListenAddr == addr {
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
