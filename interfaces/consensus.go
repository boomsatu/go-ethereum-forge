
package interfaces

import (
	"math/big"
)

// BlockHeader represents a minimal block header interface for consensus
type BlockHeader interface {
	GetNumber() uint64
	GetParentHash() [32]byte
	GetTimestamp() int64
	GetDifficulty() *big.Int
	SetDifficulty(*big.Int)
	GetHash() [32]byte
	SetHash([32]byte)
	GetNonce() uint64
	SetNonce(uint64)
}

// Block represents a minimal block interface for consensus
type Block interface {
	GetHeader() BlockHeader
	GetTransactions() []interface{}
	CalculateHash() [32]byte
}

// Engine represents the consensus engine interface
type Engine interface {
	MineBlock(block Block) error
	ValidateProofOfWork(block Block) bool
	CalculateDifficulty(currentBlock Block, parentBlock Block) *big.Int
}
