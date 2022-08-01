package main

import (
	"fmt"
	"github.com/manishmeganathan/essensio/core"
	"log"
<<<<<<< Updated upstream
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/manishmeganathan/essensio/jsonrpc"
)
=======
	"os"
)

func main() {
	// Check if a command has been entered
	if len(os.Args) < 2 {
		log.Fatalln("Command Not Found")
	}

	// Read the command and check its value
	switch cmd := os.Args[1]; cmd {
	// AddBlock command
	case "addblock":
		// Check if an input has been provided for the block data
		if len(os.Args) < 3 {
			log.Fatalln("Missing Input for 'AddBlock' command")
		}

		// Load up the BlockChain
		chain, err := core.NewChainManager()
		if err != nil {
			log.Fatalln("Failed to Start Blockchain:", err)
		}

		defer chain.Stop()

		// Add the Block with the given data to the chain
		if err := chain.AddBlock(os.Args[2]); err != nil {
			log.Fatalln("Failed to Add Block to Chain:", err)
		}

	// ShowChain command
	case "showchain":
		// Load up the BlockChain
		chain, err := core.NewChainManager()
		if err != nil {
			log.Fatalln("Failed to Start Blockchain:", err)
		}
>>>>>>> Stashed changes

// TODO:
// 1. RPC instead CLI -  Done
// 2. Tx Model
// 3. Tx Pool
// 4. Update the RPC

const SERVER_PORT = 8080

func main() {
	// Create a new RPC Server and register the JSON Codec
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	// Create a new JSON-RPC API for Essensio
	api := jsonrpc.NewAPI()
	defer api.Stop()

	// Register the Essensio API with the Server
	if err := server.RegisterService(api, ""); err != nil {
		log.Fatalln("Failed to Register Essensio API:", err)
	}

	// Set up a new Multiplexed Router
	router := mux.NewRouter()
	router.Handle("/rpc", server)

	// HTTP Listen & Serve
	fmt.Println("Server Starting...")
	if err := http.ListenAndServe(fmt.Sprintf(":%v", SERVER_PORT), router); err != nil {
		log.Fatalln(err)
	}
}
