
package core

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Miner struct {
	blockchain *Blockchain
	minerAddr  string
	running    bool
	mu         sync.Mutex
	stopChan   chan struct{}
}

func NewMiner(blockchain *Blockchain, minerAddr string) *Miner {
	return &Miner{
		blockchain: blockchain,
		minerAddr:  minerAddr,
		stopChan:   make(chan struct{}),
	}
}

func (m *Miner) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	fmt.Println("Starting miner...")

	for {
		select {
		case <-m.stopChan:
			return
		default:
			m.mineBlock()
		}
	}
}

func (m *Miner) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)
	fmt.Println("Miner stopped")
}

func (m *Miner) mineBlock() {
	// Get current block
	currentBlock := m.blockchain.GetCurrentBlock()
	if currentBlock == nil {
		time.Sleep(time.Second)
		return
	}

	// Get pending transactions
	mempool := m.blockchain.GetMempool()
	pendingTxs := mempool.GetPendingTransactions()

	// Limit transactions per block
	maxTxs := 100
	if len(pendingTxs) > maxTxs {
		pendingTxs = pendingTxs[:maxTxs]
	}

	// Create new block
	newBlock := NewBlock(
		currentBlock.Header.Hash,
		currentBlock.Header.Number+1,
		pendingTxs,
	)

	// Set miner reward transaction
	minerAddr := common.HexToAddress(m.minerAddr)
	rewardTx := NewTransaction(
		0, // nonce
		&minerAddr,
		big.NewInt(2e18), // 2 ETH reward
		21000, // gas limit
		big.NewInt(0), // gas price
		nil, // data
	)
	
	newBlock.Transactions = append([]*Transaction{rewardTx}, newBlock.Transactions...)

	// Mine the block
	fmt.Printf("Mining block %d with %d transactions...\n", newBlock.Header.Number, len(newBlock.Transactions))
	start := time.Now()
	
	newBlock.MineBlock(newBlock.Header.Difficulty)
	
	duration := time.Since(start)
	fmt.Printf("Block %d mined in %v! Hash: %s\n", newBlock.Header.Number, duration, newBlock.Header.Hash.Hex())

	// Add block to blockchain
	if err := m.blockchain.AddBlock(newBlock); err != nil {
		fmt.Printf("Failed to add mined block: %v\n", err)
		return
	}

	// Remove mined transactions from mempool
	for _, tx := range pendingTxs {
		mempool.RemoveTransaction(tx.Hash)
	}

	fmt.Printf("Block %d added to blockchain\n", newBlock.Header.Number)
}

func (m *Miner) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
