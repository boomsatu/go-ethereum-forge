package rpc

import (
	"blockchain-node/consensus"
	"blockchain-node/core"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type MiningAPI struct {
	blockchain *core.Blockchain
	miner      *core.Miner
	stats      *MiningStats
	mutex      sync.RWMutex
	isActive   bool
}

type MiningStats struct {
	IsActive     bool    `json:"isActive"`
	HashRate     float64 `json:"hashRate"`
	BlocksFound  int     `json:"blocksFound"`
	Difficulty   string  `json:"difficulty"`
	MinerAddress string  `json:"minerAddress"`
	StartTime    int64   `json:"startTime"`
}

func NewMiningAPI(blockchain *core.Blockchain) *MiningAPI {
	return &MiningAPI{
		blockchain: blockchain,
		stats: &MiningStats{
			IsActive:    false,
			HashRate:    0,
			BlocksFound: 0,
			Difficulty:  "1000",
		},
	}
}

func (api *MiningAPI) StartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MinerAddress string `json:"minerAddress"`
		Threads      int    `json:"threads"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.isActive {
		http.Error(w, "Mining already active", http.StatusConflict)
		return
	}

	// Start mining
	api.miner = core.NewMiner(api.blockchain, req.MinerAddress)
	api.isActive = true
	api.stats.IsActive = true
	api.stats.MinerAddress = req.MinerAddress
	api.stats.StartTime = time.Now().Unix()

	// Start mining in background
	go func() {
		api.miner.Start()
	}()

	response := map[string]interface{}{
		"success": true,
		"message": "Mining started successfully",
		"stats":   api.stats,
	}

	json.NewEncoder(w).Encode(response)
}

func (api *MiningAPI) StopHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api.mutex.Lock()
	defer api.mutex.Unlock()

	if !api.isActive {
		http.Error(w, "Mining not active", http.StatusConflict)
		return
	}

	if api.miner != nil {
		api.miner.Stop()
	}

	api.isActive = false
	api.stats.IsActive = false

	response := map[string]interface{}{
		"success": true,
		"message": "Mining stopped successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func (api *MiningAPI) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	api.mutex.RLock()
	defer api.mutex.RUnlock()

	// Update stats
	if api.isActive && time.Now().Unix()-api.stats.StartTime > 0 {
		api.stats.HashRate = float64(api.stats.BlocksFound) / float64(time.Now().Unix()-api.stats.StartTime)
	}

	json.NewEncoder(w).Encode(api.stats)
}

func (api *MiningAPI) MineBlockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MinerAddress string `json:"minerAddress"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Get pending transactions
	transactions := api.blockchain.GetMempool().GetPendingTransactions()

	// Create new block
	currentBlock := api.blockchain.GetCurrentBlock()
	var parentHash [32]byte
	var blockNumber uint64 = 0

	if currentBlock != nil {
		parentHash = currentBlock.Header.Hash
		blockNumber = currentBlock.Header.Number + 1
	}

	block := core.NewBlock(parentHash, blockNumber, transactions)

	// Mine the block using consensus
	consensusEngine := consensus.NewProofOfWork()
	if err := consensusEngine.MineBlock(block); err != nil {
		http.Error(w, "Failed to mine block: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add block to blockchain
	if err := api.blockchain.AddBlock(block); err != nil {
		http.Error(w, "Failed to add block: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update mining stats
	api.mutex.Lock()
	api.stats.BlocksFound++
	api.mutex.Unlock()

	response := map[string]interface{}{
		"blockNumber": block.Header.Number,
		"hash":        fmt.Sprintf("0x%x", block.Header.Hash),
		"success":     true,
	}

	json.NewEncoder(w).Encode(response)
}
