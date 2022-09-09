package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	kdht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
	"time"
)

/*
Step 1 : Setup the host
Step 2 : Set up the stream handlers
Step 3 : Setup the discovery mechanism
Step 4 : Facilitate communication
*/
var EssensioProtocol = protocol.ID("/essensio/network")
var EssensioPubSub = "essensio-pubsub"
var EssensioDiscovery = "join-essensio"

type Server struct {
	host                host.Host
	kadDHT              *kdht.IpfsDHT
	id                  peer.ID
	addr                multiaddr.Multiaddr
	pubSubRouter        *pubsub.PubSub
	pubsubTopicHandlers map[string]*pubsub.Topic
	discovery           *discovery.RoutingDiscovery
	peers               map[peer.ID]struct{}
}

func DefaultStreamHandler(s network.Stream) {
	log.Println("Got a new stream ", s.Protocol(), "remote addr", s.Conn().RemotePeer())
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Println("error reading buffer", err)
		}

		log.Println("Message received", str)
	}
}

func defaultPubSubHandler(msg *pubsub.Message) {

	log.Println("Received a pubsub message from", msg.ReceivedFrom)
	log.Println("Data", string(msg.Data))
}

func NewServer(listenerAddr multiaddr.Multiaddr, bootnodeAddr string) (*Server, error) {

	libp2pHost, err := createHost(listenerAddr)
	if err != nil {
		log.Println("error creating the libp2p libp2pHost", err)
		return nil, err
	}

	s := &Server{
		host:                libp2pHost,
		id:                  libp2pHost.ID(),
		addr:                libp2pHost.Addrs()[0],
		pubsubTopicHandlers: make(map[string]*pubsub.Topic),
		peers:               make(map[peer.ID]struct{}),
	}

	log.Println("Essensio network instantiated")

	log.Println("Node Address", fmt.Sprintf("%s/p2p/%s", s.addr, s.id))

	s.setStreamHandler(EssensioProtocol, DefaultStreamHandler)

	if err := s.setupPubSub(); err != nil {
		return nil, err
	}

	if err := s.Subscribe(EssensioPubSub, defaultPubSubHandler); err != nil {
		log.Println("Error subscribing to pubsub topic")

		return nil, err
	}

	if err := s.setupDHT(); err != nil {
		return nil, err
	}

	if bootnodeAddr != "" {
		log.Println("Connecting to bootnode", bootnodeAddr)
		if err := s.connectToBootNode(bootnodeAddr); err != nil {
			return nil, err
		}

	}

	time.Sleep(1 * time.Second)

	if err := s.Advertise(); err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Second)

	if err := s.InitDiscovery(EssensioDiscovery); err != nil {
		log.Println("Error initiating peer discovery")
	}

	return s, nil
}
func (s *Server) setupDHT() (err error) {
	// Create a new Kad DHT for the Server with the dht options
	s.kadDHT, err = kdht.New(context.Background(), s.host, kdht.ProtocolPrefix(protocol.ID(EssensioProtocol)), kdht.Mode(kdht.ModeServer))
	if err != nil {
		// Return the error
		return
	}

	// Bootstrap the Kad DHT and check for errors
	if err = s.kadDHT.Bootstrap(context.Background()); err != nil {
		// Return the error
		return
	}

	return nil
}
func (s *Server) Advertise() (err error) {

	s.discovery = discovery.NewRoutingDiscovery(s.kadDHT)

	discovery.Advertise(context.Background(), s.discovery, EssensioDiscovery)

	return nil
}

func createHost(listenerAddr multiaddr.Multiaddr) (host.Host, error) {

	randomness := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 256, randomness)
	if err != nil {
		log.Println("Error generating key pair", err)
		return nil, err
	}

	return libp2p.New(
		libp2p.Identity(prvKey),
		libp2p.ListenAddrs(listenerAddr),
	)
}

func (s *Server) setStreamHandler(protocolID protocol.ID, handler network.StreamHandler) {
	s.host.SetStreamHandler(protocolID, handler)
}

func (s *Server) ConnectPeer(addr multiaddr.Multiaddr) error {

	ctx := context.Background()
	peerAddrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Println("error parsing p2p address", err)
		return err
	}

	return s.host.Connect(ctx, *peerAddrInfo)
}

func (s *Server) OpenStream(addr multiaddr.Multiaddr) (*bufio.ReadWriter, error) {

	peerAddrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Panic("error parsing p2p address", err)
		return nil, err
	}

	log.Println(s.host.Network().Peers())

	stream, err := s.host.NewStream(context.Background(), peerAddrInfo.ID, EssensioProtocol)
	if err != nil {
		log.Panic(err)
		return nil, err
	}

	log.Println("Stream opened with remote peer", peerAddrInfo.ID, stream.ID())

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	return rw, nil
}

func (s *Server) setupPubSub() (err error) {
	s.pubSubRouter, err = pubsub.NewGossipSub(context.Background(), s.host)
	if err != nil {
		// Return the error
		return err
	}

	return nil
}

func (s *Server) Subscribe(topic string, handler func(msg *pubsub.Message)) error {
	topicHandler, err := s.pubSubRouter.Join(topic)
	if err != nil {
		return err
	}

	s.pubsubTopicHandlers[topic] = topicHandler

	subscribeHandler, err := topicHandler.Subscribe()
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := subscribeHandler.Next(context.Background())
			if err != nil {
				log.Panic(err)
			}

			if msg.ReceivedFrom == s.id {
				continue
			}

			handler(msg)
		}
	}()

	return nil
}

func (s *Server) Broadcast(topic string, data []byte) (err error) {
	handler, ok := s.pubsubTopicHandlers[topic]
	if !ok {
		handler, err = s.pubSubRouter.Join(topic)
		if err != nil {
			return err
		}

		s.pubsubTopicHandlers[topic] = handler
	}

	return handler.Publish(context.Background(), data)
}

func (s *Server) SendMessage(peermultiAddr multiaddr.Multiaddr, msg string) error {

	rw, err := s.OpenStream(peermultiAddr)
	if err != nil {
		log.Panic(err)
	}

	_, err = rw.WriteString(msg)
	if err != nil {
		log.Println("error writing to buffer")
	}

	if err := rw.Flush(); err != nil {
		log.Panic(err)
	}

	return nil
}

func (s *Server) SendPubSubMessage(topic string, msg interface{}) error {

	rawMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling the pubsub message", "error", err)

		return err
	}

	return s.Broadcast(topic, rawMsg)
}

//SendPubSubMessageBlock sends a pubsub message to the blockadded pubsub topic along with address of the original block
func (s *Server) SendPubSubMessageBlock(topic string, address string) error {

	rawMsg, err := json.Marshal(address)
	if err != nil {
		log.Println("Error marshalling the pubsub message", "error", err)

		return err
	}
	return s.Broadcast(topic, rawMsg)
}

func (s *Server) connectToBootNode(addr string) error {

	peermultiAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Panic("error parsing multi address")
	}

	if err := s.ConnectPeer(peermultiAddr); err != nil {
		log.Panic("Error connecting to bootnode ", err)
		return err
	}

	return nil
}

func (s *Server) InitDiscovery(discoveryTopic string) (err error) {

	go func() {
		for {
			peersChan, err := s.discovery.FindPeers(context.Background(), discoveryTopic)
			if err != nil {
				log.Println("Error finding peers")
			}

			for addr := range peersChan {
				if addr.ID != s.id {
					multiAddr, err := peer.AddrInfoToP2pAddrs(&addr)
					if err != nil {
						log.Println("error parsing multiaddr")
					}

					if _, ok := s.peers[addr.ID]; !ok {
						if err := s.ConnectPeer(multiAddr[0]); err != nil {
							log.Println("Error connecting to peer", err)
						}

						s.peers[addr.ID] = struct{}{}

						log.Println("Peer Connected", addr.ID)
					}
				}
			}

			<-time.After(5 * time.Second)
		}
	}()

	return nil
}
