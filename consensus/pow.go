
package consensus

import (
	"blockchain-node/core"
	"blockchain-node/crypto"
	"crypto/rand"
	"encoding/binary"
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
		minDifficulty: big.NewInt(1000),      // Minimum difficulty
		maxDifficulty: new(big.Int).Lsh(big.NewInt(1), 240), // Maximum difficulty
	}
}

// MineBlock performs proof of work mining on a block
func (pow *ProofOfWork) MineBlock(block *core.Block) error {
	target := pow.calculateTarget(block.Header.Difficulty)
	
	// Initialize nonce with random value to prevent mining collisions
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	block.Header.Nonce = binary.BigEndian.Uint64(randomBytes)
	
	startTime := time.Now()
	hashCount := uint64(0)
	
	for {
		// Calculate block hash
		hash := pow.calculateBlockHash(block)
		hashCount++
		
		// Check if hash meets difficulty target
		hashInt := new(big.Int).SetBytes(hash[:])
		if hashInt.Cmp(target) <= 0 {
			block.Header.Hash = hash
			
			// Log mining success
			elapsed := time.Since(startTime)
			hashRate := float64(hashCount) / elapsed.Seconds()
			
			// Mining successful
			return nil
		}
		
		// Increment nonce and continue
		block.Header.Nonce++
		
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
func (pow *ProofOfWork) ValidateProofOfWork(block *core.Block) bool {
	// Recalculate block hash
	hash := pow.calculateBlockHash(block)
	
	// Verify hash matches block header
	if hash != block.Header.Hash {
		return false
	}
	
	// Check if hash meets difficulty target
	target := pow.calculateTarget(block.Header.Difficulty)
	hashInt := new(big.Int).SetBytes(hash[:])
	
	return hashInt.Cmp(target) <= 0
}

// CalculateDifficulty calculates the difficulty for the next block
func (pow *ProofOfWork) CalculateDifficulty(currentBlock *core.Block, parentBlock *core.Block) *big.Int {
	// For genesis block or first few blocks, use minimum difficulty
	if currentBlock.Header.Number < DifficultyWindow {
		return new(big.Int).Set(pow.minDifficulty)
	}
	
	// Get the block from DifficultyWindow blocks ago
	targetNumber := currentBlock.Header.Number - DifficultyWindow
	if targetNumber < 0 {
		targetNumber = 0
	}
	
	// Calculate actual time taken for last DifficultyWindow blocks
	actualTime := time.Duration(currentBlock.Header.Timestamp-parentBlock.Header.Timestamp) * time.Second
	expectedTime := TargetBlockTime * DifficultyWindow
	
	// Calculate difficulty adjustment
	currentDifficulty := new(big.Int).Set(parentBlock.Header.Difficulty)
	
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

// calculateBlockHash calculates the hash of a block for mining
func (pow *ProofOfWork) calculateBlockHash(block *core.Block) [32]byte {
	// Create mining data by combining header fields
	data := make([]byte, 0, 256)
	
	// Add block number
	numberBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(numberBytes, block.Header.Number)
	data = append(data, numberBytes...)
	
	// Add parent hash
	data = append(data, block.Header.ParentHash[:]...)
	
	// Add timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(block.Header.Timestamp))
	data = append(data, timestampBytes...)
	
	// Add state root
	data = append(data, block.Header.StateRoot[:]...)
	
	// Add transactions root
	data = append(data, block.Header.TxHash[:]...)
	
	// Add receipts root
	data = append(data, block.Header.ReceiptHash[:]...)
	
	// Add gas limit
	gasLimitBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasLimitBytes, block.Header.GasLimit)
	data = append(data, gasLimitBytes...)
	
	// Add gas used
	gasUsedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasUsedBytes, block.Header.GasUsed)
	data = append(data, gasUsedBytes...)
	
	// Add difficulty
	data = append(data, block.Header.Difficulty.Bytes()...)
	
	// Add nonce
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, block.Header.Nonce)
	data = append(data, nonceBytes...)
	
	// Calculate SHA256 hash
	return crypto.SHA256Hash(data)
}

// Engine represents the consensus engine interface
type Engine interface {
	MineBlock(block *core.Block) error
	ValidateProofOfWork(block *core.Block) bool
	CalculateDifficulty(currentBlock *core.Block, parentBlock *core.Block) *big.Int
}

// Consensus errors
var (
	ErrMiningTimeout = fmt.Errorf("mining timeout exceeded")
	ErrInvalidProof  = fmt.Errorf("invalid proof of work")
)

// Add missing import
import "fmt"
