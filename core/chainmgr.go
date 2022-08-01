package core

import (
	"fmt"

	"github.com/manishmeganathan/essensio/common"
	"github.com/manishmeganathan/essensio/db"
)

var (
	ChainHeadKey   = []byte("state-chainhead")
	ChainHeightKey = []byte("state-chainheight")
)

// ChainManager represents a blockchain as a set of Blocks
type ChainManager struct {
	// Represents the database of blockchain data
	// This contains the state and blocks of the blockchain
	Db *db.Database

	// Represents the hash of the last Block
	Head common.Hash
	// Represents the Height of the chain. Last block Height+1
	Height int64
}

// String implements the Stringer interface for BlockChain
func (chain *ChainManager) String() string {
	return fmt.Sprintf("Chain Head: %x || Chain Height: %v", chain.Head, chain.Height)
}

// AddBlock generates and appends a Block to the chain for a given string data.
// The generated block is stored in the database. Any error that occurs is returned.
<<<<<<< Updated upstream:core/chainmgr/chainmgr.go
func (chain *ChainManager) AddBlock(txns core.Transactions) error {
	// Create a new Block with the given transactions
	block, err := core.NewBlock(txns, chain.Head, chain.Height)
	if err != nil {
		return fmt.Errorf("failed to generate block: %w", err)
	}
=======
func (chain *ChainManager) AddBlock(data []*Transaction) error {
	// Create a new Block with the given data
	block := NewBlock(data, chain.head, chain.height)
>>>>>>> Stashed changes:core/chainmgr.go

	// Serialize the Block
	blockData, err := block.Serialize()
	if err != nil {
		return fmt.Errorf("block serialize failed: %w", err)
	}

	// Add block to db
	if err := chain.Db.SetEntry(block.BlockHash.Bytes(), blockData); err != nil {
		return fmt.Errorf("block store to db failed: %w", err)
	}

	// Update the chain head with the new block hash and increment chain height
	chain.Head = block.BlockHash
	chain.Height++

	// Sync the chain state into the DB
	if err := chain.syncState(); err != nil {
		return fmt.Errorf("chain state sync failed: %w", err)
	}

	return nil
}

// NewChainManager returns a new BlockChain with an initialized
// Genesis Block with the provided genesis data.
func NewChainManager() (*ChainManager, error) {
	// Create a new ChainManager object
	chain := new(ChainManager)

	// Check if the database already exists
	if db.Exists() {
		// Load blockchain state from database
		if err := chain.load(); err != nil {
			return nil, fmt.Errorf("failed to load existing blockchain: %w", err)
		}

	} else {
		// Initialize blockchain state and database
		if err := chain.init(); err != nil {
			return nil, fmt.Errorf("failed to initialize new blockchain: $%w", err)
		}
	}

	return chain, nil
}

// load restarts a ChainManager from the database.
// It updates its in-memory chain state chain information from the DB.
func (chain *ChainManager) load() (err error) {
	// Open the database
	if chain.Db, err = db.Open(); err != nil {
		return err
	}

	// Get the chain head and set it
	head, err := chain.Db.GetEntry(ChainHeadKey)
	if err != nil {
		return fmt.Errorf("chain head retrieve failed: %w", err)
	}

	// Get the chain height
	height, err := chain.Db.GetEntry(ChainHeightKey)
	if err != nil {
		return fmt.Errorf("chain height retrieve failed: %w", err)
	}

	// Deserialize the height into an int64
	object, err := common.GobDecode(height, new(int64))
	if err != nil {
		return fmt.Errorf("error deserializing chain height: %w", err)
	}

	// Cast the object into an int64 and set it
	chain.Height = *object.(*int64)
	// Convert the head bytes into a Hash and set it
	chain.Head = common.BytesToHash(head)

	return nil
}

// init initializes a new chain in the database.
// It generates a Genesis Block and adds it to DB and updates all chain state data.
func (chain *ChainManager) init() (err error) {
	// Open the database
	if chain.Db, err = db.Open(); err != nil {
		return err
	}

	fmt.Println(">>>> New Blockchain Initialization. Creating Genesis Block <<<<")

<<<<<<< Updated upstream:core/chainmgr/chainmgr.go
	// Create Genesis Block
	genesisBlock, err := core.GenesisBlock()
	if err != nil {
		return fmt.Errorf("genesis block generation failed: %w", err)
	}

	// Serialize the Genesis Block
=======
	// Create Genesis Block & serialize it
	var tx []*Transaction
	genesisBlock := NewBlock(tx, common.NullHash(), 0)
>>>>>>> Stashed changes:core/chainmgr.go
	genesisData, err := genesisBlock.Serialize()
	if err != nil {
		return fmt.Errorf("block serialize failed: %w", err)
	}

	// Add Genesis Block to DB
	if err := chain.Db.SetEntry(genesisBlock.BlockHash.Bytes(), genesisData); err != nil {
		return fmt.Errorf("genesis block store to db failed: %w", err)
	}

	// Set the chain height and head into struct
	chain.Head, chain.Height = genesisBlock.BlockHash, 1

	// Sync the chain state into the DB
	if err := chain.syncState(); err != nil {
		return fmt.Errorf("chain state sync failed: %w", err)
	}

	return nil
}

// Stop closes the ChainManager's database client
func (chain *ChainManager) Stop() {
	chain.Db.Close()
}

// syncState updates the chain head and height values into the DB at keys
// specified by the ChainHeadKey and ChainHeightKey respectively.
func (chain *ChainManager) syncState() error {
	// Sync chain head into the DB
<<<<<<< Updated upstream:core/chainmgr/chainmgr.go
	if err := chain.db.SetEntry(ChainHeadKey, chain.Head.Bytes()); err != nil {
=======
	if err := chain.Db.SetEntry(ChainHeadKey, chain.head.Bytes()); err != nil {
>>>>>>> Stashed changes:core/chainmgr.go
		return fmt.Errorf("error syncing chain head: %w", err)
	}

	// Serialize the chain height
	height, err := common.GobEncode(chain.Height)
	if err != nil {
		return fmt.Errorf("error serializing chain height: %w", err)
	}

	// Sync the encoded height into the DB
	if err := chain.Db.SetEntry(ChainHeightKey, height); err != nil {
		return fmt.Errorf("error syncing chain height: %w", err)
	}

	return nil
}
