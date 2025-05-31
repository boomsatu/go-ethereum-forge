package rpc

import (
	"blockchain-node/core"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	config     *Config
	blockchain *core.Blockchain
	server     *http.Server
	walletAPI  *WalletAPI
}

func NewServer(config *Config, blockchain *core.Blockchain) *Server {
	return &Server{
		config:     config,
		blockchain: blockchain,
		walletAPI:  NewWalletAPI(blockchain),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// JSON-RPC endpoint
	mux.HandleFunc("/", s.handleRPC)
	
	// Wallet API endpoints
	mux.HandleFunc("/api/wallet/create", s.walletAPI.CreateHandler)
	mux.HandleFunc("/api/wallet/import", s.walletAPI.ImportHandler)
	mux.HandleFunc("/api/wallet/send", s.walletAPI.SendTransactionHandler)
	mux.HandleFunc("/api/wallet/balance", s.walletAPI.CheckBalanceHandler)
	
	// Admin API endpoints
	mux.HandleFunc("/api/admin/status", s.handleAdminStatus)
	mux.HandleFunc("/api/admin/start", s.handleAdminStart)
	mux.HandleFunc("/api/admin/stop", s.handleAdminStop)
	
	// Mining API endpoints
	mux.HandleFunc("/api/mining/start", s.handleMiningStart)
	mux.HandleFunc("/api/mining/stop", s.handleMiningStop)
	mux.HandleFunc("/api/mining/stats", s.handleMiningStats)
	mux.HandleFunc("/api/mining/mine-block", s.handleMineBlock)
	
	// Network API endpoints
	mux.HandleFunc("/api/network/stats", s.handleNetworkStats)
	mux.HandleFunc("/api/network/peers", s.handleNetworkPeers)
	
	// Metrics endpoint
	mux.HandleFunc("/api/metrics", s.handleMetrics)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler: corsMiddleware(mux),
	}

	log.Printf("RPC server starting on %s:%d", s.config.Host, s.config.Port)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JsonRPC string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
		ID      interface{}   `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	var result interface{}
	var rpcErr *RPCError

	switch req.Method {
	case "eth_chainId":
		result = fmt.Sprintf("0x%x", s.blockchain.GetChainID())
	case "net_version":
		result = strconv.FormatUint(s.blockchain.GetChainID(), 10)
	case "eth_blockNumber":
		if currentBlock := s.blockchain.GetCurrentBlock(); currentBlock != nil {
			result = fmt.Sprintf("0x%x", currentBlock.Header.Number)
		} else {
			result = "0x0"
		}
	case "eth_getBalance":
		result, rpcErr = s.handleGetBalance(req.Params)
	case "eth_getTransactionCount":
		result, rpcErr = s.handleGetTransactionCount(req.Params)
	case "eth_getBlockByNumber":
		result, rpcErr = s.handleGetBlockByNumber(req.Params)
	case "eth_getBlockByHash":
		result, rpcErr = s.handleGetBlockByHash(req.Params)
	case "eth_getTransactionByHash":
		result, rpcErr = s.handleGetTransactionByHash(req.Params)
	case "eth_getTransactionReceipt":
		result, rpcErr = s.handleGetTransactionReceipt(req.Params)
	case "eth_sendTransaction":
		result, rpcErr = s.handleSendTransaction(req.Params)
	case "eth_sendRawTransaction":
		result, rpcErr = s.handleSendRawTransaction(req.Params)
	default:
		rpcErr = &RPCError{Code: -32601, Message: "Method not found"}
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      req.ID,
	}

	if rpcErr != nil {
		response["error"] = rpcErr
	} else {
		response["result"] = result
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handleGetBalance(params []interface{}) (interface{}, *RPCError) {
	if len(params) < 1 {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	addressStr, ok := params[0].(string)
	if !ok {
		return nil, &RPCError{Code: -32602, Message: "Invalid address parameter"}
	}

	// Clean address
	addressStr = strings.TrimSpace(addressStr)
	if strings.HasPrefix(addressStr, "0x") {
		addressStr = addressStr[2:]
	}

	if len(addressStr) != 40 {
		return nil, &RPCError{Code: -32602, Message: "Invalid address format"}
	}

	var address [20]byte
	for i := 0; i < 20; i++ {
		fmt.Sscanf(addressStr[i*2:i*2+2], "%02x", &address[i])
	}

	balance := s.blockchain.GetStateDB().GetBalance(address)
	return fmt.Sprintf("0x%x", balance), nil
}

func (s *Server) handleGetTransactionCount(params []interface{}) (interface{}, *RPCError) {
	if len(params) < 1 {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	addressStr, ok := params[0].(string)
	if !ok {
		return nil, &RPCError{Code: -32602, Message: "Invalid address parameter"}
	}

	// Clean address
	addressStr = strings.TrimSpace(addressStr)
	if strings.HasPrefix(addressStr, "0x") {
		addressStr = addressStr[2:]
	}

	var address [20]byte
	for i := 0; i < 20; i++ {
		fmt.Sscanf(addressStr[i*2:i*2+2], "%02x", &address[i])
	}

	nonce := s.blockchain.GetStateDB().GetNonce(address)
	return fmt.Sprintf("0x%x", nonce), nil
}

func (s *Server) handleGetBlockByNumber(params []interface{}) (interface{}, *RPCError) {
	if len(params) < 2 {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	blockNumStr, ok := params[0].(string)
	if !ok {
		return nil, &RPCError{Code: -32602, Message: "Invalid block number parameter"}
	}

	var blockNum uint64
	if blockNumStr == "latest" {
		if currentBlock := s.blockchain.GetCurrentBlock(); currentBlock != nil {
			blockNum = currentBlock.Header.Number
		} else {
			return nil, nil
		}
	} else {
		if strings.HasPrefix(blockNumStr, "0x") {
			blockNumStr = blockNumStr[2:]
		}
		var err error
		blockNum, err = strconv.ParseUint(blockNumStr, 16, 64)
		if err != nil {
			return nil, &RPCError{Code: -32602, Message: "Invalid block number format"}
		}
	}

	block := s.blockchain.GetBlockByNumber(blockNum)
	if block == nil {
		return nil, nil
	}

	return s.formatBlock(block), nil
}

func (s *Server) handleGetBlockByHash(params []interface{}) (interface{}, *RPCError) {
	if len(params) < 2 {
		return nil, &RPCError{Code: -32602, Message: "Invalid params"}
	}

	hashStr, ok := params[0].(string)
	if !ok {
		return nil, &RPCError{Code: -32602, Message: "Invalid hash parameter"}
	}

	// Clean hash
	if strings.HasPrefix(hashStr, "0x") {
		hashStr = hashStr[2:]
	}

	var hash [32]byte
	for i := 0; i < 32 && i*2 < len(hashStr); i++ {
		fmt.Sscanf(hashStr[i*2:i*2+2], "%02x", &hash[i])
	}

	block := s.blockchain.GetBlockByHash(hash)
	if block == nil {
		return nil, nil
	}

	return s.formatBlock(block), nil
}

func (s *Server) formatBlock(block *core.Block) map[string]interface{} {
	txHashes := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txHashes[i] = fmt.Sprintf("0x%x", tx.Hash)
	}

	return map[string]interface{}{
		"number":           fmt.Sprintf("0x%x", block.Header.Number),
		"hash":             fmt.Sprintf("0x%x", block.Header.Hash),
		"parentHash":       fmt.Sprintf("0x%x", block.Header.ParentHash),
		"timestamp":        fmt.Sprintf("0x%x", block.Header.Timestamp),
		"gasLimit":         fmt.Sprintf("0x%x", block.Header.GasLimit),
		"gasUsed":          fmt.Sprintf("0x%x", block.Header.GasUsed),
		"difficulty":       fmt.Sprintf("0x%x", block.Header.Difficulty),
		"transactionCount": len(block.Transactions),
		"transactions":     txHashes,
		"size":            "0x0",
		"miner":           "0x0000000000000000000000000000000000000000",
	}
}

func (s *Server) handleGetTransactionByHash(params []interface{}) (interface{}, *RPCError) {
	// Implementation for getting transaction by hash
	return nil, &RPCError{Code: -32601, Message: "Not implemented"}
}

func (s *Server) handleGetTransactionReceipt(params []interface{}) (interface{}, *RPCError) {
	// Implementation for getting transaction receipt
	return nil, &RPCError{Code: -32601, Message: "Not implemented"}
}

func (s *Server) handleSendTransaction(params []interface{}) (interface{}, *RPCError) {
	// Implementation for sending transaction
	return nil, &RPCError{Code: -32601, Message: "Not implemented"}
}

func (s *Server) handleSendRawTransaction(params []interface{}) (interface{}, *RPCError) {
	// Implementation for sending raw transaction
	return nil, &RPCError{Code: -32601, Message: "Not implemented"}
}

func (s *Server) handleAdminStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status": "running",
		"config": map[string]interface{}{
			"chainId":   s.blockchain.GetChainID(),
			"dataDir":   s.blockchain.GetConfig().DataDir,
			"gasLimit":  s.blockchain.GetConfig().BlockGasLimit,
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAdminStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleAdminStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMiningStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMiningStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *Server) handleMiningStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := map[string]interface{}{
		"isActive":    false,
		"hashRate":    0,
		"blocksFound": 0,
		"difficulty":  "1024",
	}
	
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleMineBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blockNumber": 1,
		"hash":        "0x0",
	})
}

func (s *Server) handleNetworkStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := map[string]interface{}{
		"peerCount":  0,
		"difficulty": "1024",
		"hashRate":   "0",
	}
	
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleNetworkPeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	metrics := map[string]interface{}{
		"uptime":            time.Now().Unix(),
		"memoryUsage":       100 * 1024 * 1024,
		"diskUsage":         500 * 1024 * 1024,
		"cpuUsage":          10.5,
		"blockCount":        1,
		"transactionCount":  0,
		"peersConnected":    0,
	}
	
	json.NewEncoder(w).Encode(metrics)
}
