/*
This module contains the utility function that are used
for various hashing functionality
*/
package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// A type alias for a byte slice that represents a hash
type Hash []byte

// A type alias for a byte slice that represents a wallet public key
type PublicKey []byte

func Hash256(payload []byte) Hash {
	// Declare the hash byte array
	var hash [32]byte

	// Generate the hash of the payload
	hash = sha3.Sum256(payload)
	// Generate the hash of the hash
	hash = sha3.Sum256(hash[:])

	// Return the hash as slice
	return hash[:]
}

/*
A function that generates a 160bit hash output for a given slice of bytes payload.
Generated by hashing the payload with 256 bit hash algorithm, then a 224 bit
algorithm and finally truncated to 160bits.
This is equivalent in terms of security to the Bitcoin method of using SHA256 followed by RIPEMD160.
The RIPEMD160 algorithm was rejected because of its deprecated nature in the Golang crypto library
along with the fact that truncated SHA3 algortithms are equally safe as their full length versions
Hence the choice to use truncated SHA3-224 over the SHA1 which outputs 20 byte hashes directly.
hash160 = first 160 bits of [SHA3-224(SHA3-256(payload))]
References:
https://bitcoin.stackexchange.com/questions/16543/why-was-the-ripemd-160-hash-algorithms-chosen-before-sha-1
https://crypto.stackexchange.com/questions/3153/sha-256-vs-any-256-bits-of-sha-512-which-is-more-secure
*/
func Hash160(payload []byte) Hash {
	// Generate the 256bit hash of the payload
	hash256 := sha3.Sum256(payload)
	// Generate the 224bit hash of the 256bit hash
	hash224 := sha3.Sum224(hash256[:])

	// Truncate the 224bit hash to 160bit (20bytes)
	hash := hash224[:20]

	// Return the hash
	return hash
}

/*
A function that generates a 32bit hash output for a given slice of bytes payload.
Generated by double hashing the payload with a 256bit hash algorithm
and then truncating to 32bits.
This hash is used to represents the 32bit checksum of a payload.
hash32 = first 32 bits of [SHA3-256(SHA3-256(payload))]
*/
func Hash32(payload []byte) Hash {
	// Generate the 256 bit hash of the payload
	hash := Hash256(payload)
	// Truncate the hash to 32bits (4bytes)
	checksum := hash[:4]

	// Return the checksum
	return checksum
}

/*
A function that generates a pair of digital
signature keys based on the ECDSA algorithm
using the secp256r1 elliptic curve.
This method of generating cryptographic key pairs
creates a pair with 1 in 10^77 chance of collision.
*/
func KeyGenECDSA() (ecdsa.PrivateKey, PublicKey) {
	// Create a sepc256r1 elliptical curve
	curve := elliptic.P256()

	// Generate a set of keys with ECDSA algorithm
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		// Log a fatal error
		logrus.WithFields(logrus.Fields{"error": err}).Errorln("failed to generate an ECDSA key pair.")
	}

	// Construct the public key by appending the Y coordinate bytes into the X coordinate slice
	public := append(key.PublicKey.X.Bytes(), key.PublicKey.Y.Bytes()...)

	// Return private and public keys
	return *key, public
}
