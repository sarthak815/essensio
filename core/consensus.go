package core

import (
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/manishmeganathan/essensio/common"
)

// Difficulty represents the number of bits that need to be 0 for the Proof Of Work Algorithm.
// Currently Static, but can eventually be adjusted based on the total hash rate of the network,
// to achieve a block time of n minutes.
const Difficulty uint8 = 18

// GenerateTarget returns a big.Int with the target hash value for the current difficulty
func GenerateTarget() *big.Int {
	// Generate a new big Integer and left shift to match difficulty
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(Difficulty))

	return target
}

// Mint is the Proof of Work routine that generates a nonce
// that is valid for the Target difficulty of the header.
func (header *BlockHeader) Mint(height int64) common.Hash {
	var hash []byte
	var bigHash big.Int

	// Reset Nonce
	header.Nonce = 0

	for header.Nonce < math.MaxInt64 {
		// Serialize the Header
		data, err := common.GobEncode(header)
		if err != nil {
			log.Fatalln("Gob Encode for BlockHeader failed during PoW:", err)
		}

		// Hash the Header data
		hash = common.Hash256(data)
		bigHash.SetBytes(hash)

		fmt.Printf("\rBlock %v: %x", height, hash)

		// Compare the hash with target
		if bigHash.Cmp(header.Target) == -1 {
			// Block Mined!
			fmt.Printf(" [%v]\n", header.Nonce)

			break
		} else {
			// Increment Nonce & Repeat
			header.Nonce++
		}
	}

	return hash
}

// Validate is the Proof of Work validation routine.
// Returns a boolean indicating if the hash of the block is valid for its target.
func (header *BlockHeader) Validate() bool {
	var bigHash big.Int

	// Serialize the Header
	data, err := common.GobEncode(header)
	if err != nil {
		log.Fatalln("Gob Encode for BlockHeader failed during PoW:", err)
	}

	// Hash the Header data
	hash := common.Hash256(data)
	bigHash.SetBytes(hash)

	// Compare hash with target
	return bigHash.Cmp(header.Target) == -1
}