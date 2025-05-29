
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/crypto/sha3"
)

var (
	MaxTarget = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
)

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	
	return privateKey, &privateKey.PublicKey, nil
}

// GenerateEthKeyPair generates a new Ethereum-compatible ECDSA key pair
func GenerateEthKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	
	return privateKey, &privateKey.PublicKey, nil
}

// SHA256Hash calculates SHA256 hash
func SHA256Hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// Keccak256Hash calculates Keccak256 hash (Ethereum style)
func Keccak256Hash(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}

// Keccak256 calculates Keccak256 hash and returns bytes
func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Sign signs data with private key
func Sign(hash []byte, privateKey []byte) ([]byte, error) {
	return crypto.Sign(hash, privateKey)
}

// Ecrecover recovers public key from signature
func Ecrecover(hash []byte, signature []byte) (*ecdsa.PublicKey, error) {
	pubKeyBytes, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return nil, err
	}
	
	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, err
	}
	
	return pubKey, nil
}

// PubkeyToAddress converts public key to Ethereum address
func PubkeyToAddress(pubKey *ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(*pubKey)
}

// PrivateKeyToAddress converts private key to Ethereum address
func PrivateKeyToAddress(privateKey *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(privateKey.PublicKey)
}

// ValidateProofOfWork validates proof of work
func ValidateProofOfWork(hash common.Hash, nonce uint64, difficulty *big.Int) bool {
	target := new(big.Int).Div(MaxTarget, difficulty)
	hashInt := new(big.Int).SetBytes(hash[:])
	return hashInt.Cmp(target) == -1
}

// VerifySignature verifies ECDSA signature
func VerifySignature(pubKey *ecdsa.PublicKey, hash []byte, signature []byte) bool {
	if len(signature) != 65 {
		return false
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	
	return ecdsa.Verify(pubKey, hash, r, s)
}

// RecoverAddress recovers address from signature
func RecoverAddress(hash []byte, signature []byte) (common.Address, error) {
	if len(signature) != 65 {
		return common.Address{}, errors.New("invalid signature length")
	}
	
	pubKeyBytes, err := secp256k1.RecoverPubkey(hash, signature)
	if err != nil {
		return common.Address{}, err
	}
	
	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return common.Address{}, err
	}
	
	return crypto.PubkeyToAddress(*pubKey), nil
}
