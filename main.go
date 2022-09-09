package main

import (
	"flag"
	"fmt"
	"github.com/essensio_network/core/chainmgr"
	"github.com/essensio_network/core/txpool"
	"github.com/essensio_network/network"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/essensio_network/jsonrpc"
)

// TODO:
// 1. RPC instead CLI -  Done
// 2. Tx Model
// 3. Tx Pool
// 4. Update the RPC

const SERVER_PORT = 8080

var TransactionPubsubTopic = "transaction"
var TransactionPubsubTopic2 = "blockadded"
var chHeight int64 = 1

func main() {

	port := flag.Int("port", 1604, "Listener port")
	rpcPort := flag.Int("rpc-port", SERVER_PORT, "Listener port")
	ipAddr := flag.String("ip", "0.0.0.0", "Ip Address")
	bootNodeAddr := flag.String("bootnode", "", "MultiAddress of the bootnode")
	peer := flag.String("peer", "", "MultiAddr of the peer")
	flag.Parse()
	// Create a new RPC Server and register the JSON Codec
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	chain, err := chainmgr.NewChainManager()
	if err != nil {
		log.Fatalln("Failed to Start Blockchain:", err)
	}

	pool := txpool.NewTxnNoncePool()

	// Create a new JSON-RPC API for Essensio
	api := jsonrpc.NewAPI(chain, pool)
	defer api.Stop()

	// Register the Essensio API with the Server
	if err := server.RegisterService(api, "essensio"); err != nil {
		log.Fatalln("Failed to Register Essensio API:", err)
	}

	// Set up a new Multiplexed Router
	router := mux.NewRouter()
	router.Handle("/rpc", server)

	addr := fmt.Sprintf("/ip4/%s/tcp/%d", *ipAddr, *port)

	multiAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Panic("error parsing multi address")
	}

	networkServer, err := network.NewServer(multiAddr, *bootNodeAddr)
	if err != nil {
		log.Panic("error creating network server")
	}

	if *peer != "" {
		peermultiAddr, err := multiaddr.NewMultiaddr(*peer)
		if err != nil {
			log.Panic("error parsing multi address")
		}

		if err := networkServer.ConnectPeer(peermultiAddr); err != nil {
			log.Panic("Error connecting to remote peer ", err)
		}

		log.Println("Peer Connected ", peermultiAddr.String())

	}

	go func() {
		for {
			select {
			case tx := <-pool.NewTransactionsChan():
				log.Println("Sending transaction message")
				if err := networkServer.SendPubSubMessage(TransactionPubsubTopic, tx); err != nil {
					log.Println("error sending transaction")
				}
			default:
			}
		}
	}()
	/* The following go routine keeps checking for addition of a block to the chain by comparing the chain length, in case
	   of an increase in chain height the pubsub message is triggered launching a pubsub message to all other nodes available*/
	go func() {
		for {
			if chHeight < chain.Height {
				log.Println("Sending transaction message")
				if err := networkServer.SendPubSubMessageBlock(TransactionPubsubTopic2, addr); err != nil {
					log.Println("error sending transaction")
				}
				chHeight++
			}
		}
	}()

	if err := networkServer.Subscribe(TransactionPubsubTopic, pool.PubSubHandler); err != nil {
		log.Println("pubsub subscription failed")
	}
	//TransactionPubsubTopic2 is created to check if block is added and subscribes to the channel
	if err := networkServer.Subscribe(TransactionPubsubTopic2, PubSubHandlerBlock); err != nil {
		log.Println("pubsub subscription failed")
	}

	// HTTP Listen & Serve
	fmt.Println("Starting http server...")
	if err := http.ListenAndServe(fmt.Sprintf(":%v", *rpcPort), router); err != nil {
		log.Fatalln(err)
	}
}

//PubSubHandlerBlock is used to display the chain in which the new block has been added
func PubSubHandlerBlock(msg *pubsub.Message) {
	log.Println("Block added in chain: ", msg.ReceivedFrom)

}
