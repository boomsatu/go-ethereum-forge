
package consensus

import (
	"blockchain-node/crypto"
	"blockchain-node/interfaces"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Difficulty adjustment parameters
const (
	TargetBlockTime    = 15 * time.Second // Target 15 seconds per block
	DifficultyWindow   = 10               // Adjust difficulty every 10 blocks
	MaxDifficultyShift = 4                // Maximum 4x difficulty change
)

// ProofOfWork implements custom Proof of Work consensus algorithm
type ProofOfWork struct {
	minDifficulty *big.Int
	maxDifficulty *big.Int
}

// NewProofOfWork creates a new PoW consensus engine
func NewProofOfWork() *ProofOfWork {
	return &ProofOfWork{
		minDifficulty: big.NewInt(1000),                          // Minimum difficulty
		maxDifficulty: new(big.Int).Lsh(big.NewInt(1), 240),     // Maximum difficulty
	}
}

// MineBlock performs proof of work mining on a block
func (pow *ProofOfWork) MineBlock(block interfaces.Block) error {
	header := block.GetHeader()
	target := pow.calculateTarget(header.GetDifficulty())
	
	// Initialize nonce with random value to prevent mining collisions
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	header.SetNonce(binary.BigEndian.Uint64(randomBytes))
	
	startTime := time.Now()
	hashCount := uint64(0)
	
	for {
		// Calculate block hash
		hash := block.CalculateHash()
		hashCount++
		
		// Check if hash meets difficulty target
		hashInt := new(big.Int).SetBytes(hash[:])
		if hashInt.Cmp(target) <= 0 {
			header.SetHash(hash)
			return nil
		}
		
		// Increment nonce and continue
		header.SetNonce(header.GetNonce() + 1)
		
		// Prevent infinite loop - check every 100k iterations
		if hashCount%100000 == 0 {
			elapsed := time.Since(startTime)
			if elapsed > 5*time.Minute { // Max 5 minutes mining
				return ErrMiningTimeout
			}
		}
	}
}

// ValidateProofOfWork validates the proof of work for a block
func (pow *ProofOfWork) ValidateProofOfWork(block interfaces.Block) bool {
	header := block.GetHeader()
	
	// Recalculate block hash
	hash := block.CalculateHash()
	
	// Verify hash matches block header
	if hash != header.GetHash() {
		return false
	}
	
	// Check if hash meets difficulty target
	target := pow.calculateTarget(header.GetDifficulty())
	hashInt := new(big.Int).SetBytes(hash[:])
	
	return hashInt.Cmp(target) <= 0
}

// CalculateDifficulty calculates the difficulty for the next block
func (pow *ProofOfWork) CalculateDifficulty(currentBlock interfaces.Block, parentBlock interfaces.Block) *big.Int {
	currentHeader := currentBlock.GetHeader()
	parentHeader := parentBlock.GetHeader()
	
	// For genesis block or first few blocks, use minimum difficulty
	if currentHeader.GetNumber() < DifficultyWindow {
		return new(big.Int).Set(pow.minDifficulty)
	}
	
	// Calculate actual time taken for last DifficultyWindow blocks
	actualTime := time.Duration(currentHeader.GetTimestamp()-parentHeader.GetTimestamp()) * time.Second
	expectedTime := TargetBlockTime * DifficultyWindow
	
	// Calculate difficulty adjustment
	currentDifficulty := new(big.Int).Set(parentHeader.GetDifficulty())
	
	// If blocks are coming too fast, increase difficulty
	if actualTime < expectedTime/2 {
		// Increase difficulty by at most MaxDifficultyShift
		adjustment := new(big.Int).Div(currentDifficulty, big.NewInt(MaxDifficultyShift))
		currentDifficulty.Add(currentDifficulty, adjustment)
	} else if actualTime > expectedTime*2 {
		// Decrease difficulty by at most MaxDifficultyShift
		adjustment := new(big.Int).Div(currentDifficulty, big.NewInt(MaxDifficultyShift))
		currentDifficulty.Sub(currentDifficulty, adjustment)
	}
	
	// Ensure difficulty stays within bounds
	if currentDifficulty.Cmp(pow.minDifficulty) < 0 {
		currentDifficulty.Set(pow.minDifficulty)
	}
	if currentDifficulty.Cmp(pow.maxDifficulty) > 0 {
		currentDifficulty.Set(pow.maxDifficulty)
	}
	
	return currentDifficulty
}

// calculateTarget calculates the target hash value for given difficulty
func (pow *ProofOfWork) calculateTarget(difficulty *big.Int) *big.Int {
	// Target = 2^256 / difficulty
	target := new(big.Int).Div(crypto.MaxTarget, difficulty)
	return target
}

// Consensus errors
var (
	ErrMiningTimeout = errors.New("mining timeout exceeded")
	ErrInvalidProof  = errors.New("invalid proof of work")
)
