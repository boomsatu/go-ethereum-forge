
package rpc

import (
	"blockchain-node/core"
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type NetworkAPI struct {
	blockchain *core.Blockchain
}

func NewNetworkAPI(blockchain *core.Blockchain) *NetworkAPI {
	return &NetworkAPI{
		blockchain: blockchain,
	}
}

func (api *NetworkAPI) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	currentBlock := api.blockchain.GetCurrentBlock()
	blockHeight := uint64(0)
	if currentBlock != nil {
		blockHeight = currentBlock.Header.Number
	}

	stats := map[string]interface{}{
		"peerCount":   0, // Will be updated by P2P server
		"blockHeight": blockHeight,
		"difficulty":  "1000",
		"hashRate":    "0",
		"chainId":     api.blockchain.GetConfig().ChainID,
		"syncStatus": map[string]interface{}{
			"isSyncing":     false,
			"currentBlock":  blockHeight,
			"highestBlock":  blockHeight,
		},
	}

	json.NewEncoder(w).Encode(stats)
}

func (api *NetworkAPI) PeersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Return empty peers list for now
	// This will be populated by the P2P server
	peers := []map[string]interface{}{}

	json.NewEncoder(w).Encode(peers)
}

func (api *NetworkAPI) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentBlock := api.blockchain.GetCurrentBlock()
	blockCount := uint64(0)
	if currentBlock != nil {
		blockCount = currentBlock.Header.Number + 1
	}

	// Count transactions
	transactionCount := uint64(0)
	for i := uint64(0); i < blockCount; i++ {
		block := api.blockchain.GetBlockByNumber(i)
		if block != nil {
			transactionCount += uint64(len(block.Transactions))
		}
	}

	metrics := map[string]interface{}{
		"uptime":           time.Now().Unix(),
		"memoryUsage":      m.Alloc,
		"diskUsage":        0, // Placeholder
		"cpuUsage":         0, // Placeholder
		"blockCount":       blockCount,
		"transactionCount": transactionCount,
		"peersConnected":   0, // Will be updated by P2P server
		"gasUsed":          0, // Placeholder
		"gasLimit":         api.blockchain.GetConfig().BlockGasLimit,
		"pendingTxs":       len(api.blockchain.GetMempool().GetPendingTransactions()),
	}

	json.NewEncoder(w).Encode(metrics)
}
