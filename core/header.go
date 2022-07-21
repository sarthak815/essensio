package core

import (
	"math/big"
	"time"

	"github.com/manishmeganathan/essensio/common"
)

// BlockHeader is a struct that contains all the fields
// of the block that are relevant to its cryptographic integrity.
// The Block Hash is the hash of the Block Header.
type BlockHeader struct {
	// Hash of the previous block
	Priori common.Hash
	// Hash of the all the data in the block
	Summary common.Hash
	// Timestamp at the time of block creation
	Timestamp int64

	// Proof of Work Target Hash
	Target *big.Int
	// Proof of Work Nonce
	Nonce int64
}

// NewBlockHeader returns a new BlockHeader for a given priori and summary hash
func NewBlockHeader(priori, summary common.Hash) BlockHeader {
	return BlockHeader{
		priori,
		summary,
		time.Now().Unix(),
		GenerateTarget(),
		0,
	}
}