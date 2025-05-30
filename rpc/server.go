package rpc

import (
	"blockchain-node/core"
	"blockchain-node/security"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	config     *Config
	blockchain *core.Blockchain
	security   *security.SecurityManager
	server     *http.Server
	adminAPI   *AdminAPI
	miningAPI  *MiningAPI
	walletAPI  *WalletAPI
	networkAPI *NetworkAPI
}

type JSONRPCRequest struct {
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  []interface{} `json:"params"`
	Version string      `json:"jsonrpc"`
}

type JSONRPCResponse struct {
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	Version string      `json:"jsonrpc"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewServer(config *Config, blockchain *core.Blockchain) *Server {
	server := &Server{
		config:     config,
		blockchain: blockchain,
		security:   security.NewSecurityManager(),
		adminAPI:   NewAdminAPI(blockchain),
		miningAPI:  NewMiningAPI(blockchain),
		walletAPI:  NewWalletAPI(blockchain),
		networkAPI: NewNetworkAPI(blockchain),
	}
	return server
}

func (s *Server) Start() error {
	router := mux.NewRouter()
	
	// JSON-RPC endpoint
	router.HandleFunc("/", s.handleRPC).Methods("POST")
	router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// REST API endpoints for React app
	api := router.PathPrefix("/api").Subrouter()
	
	// Admin endpoints
	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/start", s.adminAPI.StartHandler).Methods("POST", "OPTIONS")
	admin.HandleFunc("/stop", s.adminAPI.StopHandler).Methods("POST", "OPTIONS")
	admin.HandleFunc("/status", s.adminAPI.StatusHandler).Methods("GET", "OPTIONS")
	admin.HandleFunc("/config", s.adminAPI.ConfigHandler).Methods("POST", "OPTIONS")

	// Mining endpoints
	mining := api.PathPrefix("/mining").Subrouter()
	mining.HandleFunc("/start", s.miningAPI.StartHandler).Methods("POST", "OPTIONS")
	mining.HandleFunc("/stop", s.miningAPI.StopHandler).Methods("POST", "OPTIONS")
	mining.HandleFunc("/stats", s.miningAPI.StatsHandler).Methods("GET", "OPTIONS")
	mining.HandleFunc("/mine-block", s.miningAPI.MineBlockHandler).Methods("POST", "OPTIONS")

	// Wallet endpoints
	walletRouter := api.PathPrefix("/wallet").Subrouter()
	walletRouter.HandleFunc("/create", s.walletAPI.CreateHandler).Methods("POST", "OPTIONS")
	walletRouter.HandleFunc("/import", s.walletAPI.ImportHandler).Methods("POST", "OPTIONS")
	walletRouter.HandleFunc("/send", s.walletAPI.SendTransactionHandler).Methods("POST", "OPTIONS")

	// Network endpoints
	network := api.PathPrefix("/network").Subrouter()
	network.HandleFunc("/stats", s.networkAPI.StatsHandler).Methods("GET", "OPTIONS")
	network.HandleFunc("/peers", s.networkAPI.PeersHandler).Methods("GET", "OPTIONS")
	
	// Metrics endpoint
	api.HandleFunc("/metrics", s.networkAPI.MetricsHandler).Methods("GET", "OPTIONS")

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("RPC server error: %v\n", err)
		}
	}()

	fmt.Printf("JSON-RPC server with REST API started on %s\n", addr)
	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Close()
		fmt.Println("JSON-RPC server stopped")
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, nil, -32700, "Parse error")
		return
	}

	result, err := s.handleMethod(req.Method, req.Params)
	if err != nil {
		s.sendError(w, req.ID, -32603, err.Error())
		return
	}

	response := JSONRPCResponse{
		ID:      req.ID,
		Result:  result,
		Version: "2.0",
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleMethod(method string, params []interface{}) (interface{}, error) {
	switch method {
	case "eth_blockNumber":
		return s.ethBlockNumber()
	case "eth_getBalance":
		return s.ethGetBalance(params)
	case "eth_getBlockByNumber":
		return s.ethGetBlockByNumber(params)
	case "eth_getBlockByHash":
		return s.ethGetBlockByHash(params)
	case "eth_getTransactionByHash":
		return s.ethGetTransactionByHash(params)
	case "eth_getTransactionReceipt":
		return s.ethGetTransactionReceipt(params)
	case "eth_sendRawTransaction":
		return s.ethSendRawTransaction(params)
	case "eth_call":
		return s.ethCall(params)
	case "eth_estimateGas":
		return s.ethEstimateGas(params)
	case "eth_gasPrice":
		return s.ethGasPrice()
	case "eth_chainId":
		return s.ethChainId()
	case "eth_getTransactionCount":
		return s.ethGetTransactionCount(params)
	case "eth_getCode":
		return s.ethGetCode(params)
	case "eth_getStorageAt":
		return s.ethGetStorageAt(params)
	case "eth_getLogs":
		return s.ethGetLogs(params)
	case "net_version":
		return s.netVersion()
	case "web3_clientVersion":
		return "blockchain-node/1.0.0", nil
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

func (s *Server) ethBlockNumber() (interface{}, error) {
	currentBlock := s.blockchain.GetCurrentBlock()
	if currentBlock == nil {
		return "0x0", nil
	}
	return fmt.Sprintf("0x%x", currentBlock.Header.Number), nil
}

func (s *Server) ethGetBalance(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	addressStr := params[0].(string)
	addressBytes, err := hex.DecodeString(strings.TrimPrefix(addressStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}
	
	var address [20]byte
	copy(address[:], addressBytes)
	
	balance := s.blockchain.GetStateDB().GetBalance(address)
	return fmt.Sprintf("0x%x", balance), nil
}

func (s *Server) ethGetBlockByNumber(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing block number parameter")
	}

	blockNumStr := params[0].(string)
	var blockNum uint64
	var err error

	if blockNumStr == "latest" {
		if block := s.blockchain.GetCurrentBlock(); block != nil {
			blockNum = block.Header.Number
		}
	} else {
		blockNum, err = strconv.ParseUint(strings.TrimPrefix(blockNumStr, "0x"), 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid block number: %v", err)
		}
	}

	block := s.blockchain.GetBlockByNumber(blockNum)
	if block == nil {
		return nil, nil
	}

	fullTx := len(params) > 1 && params[1].(bool)
	return s.formatBlock(block, fullTx), nil
}

func (s *Server) ethGetBlockByHash(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing block hash parameter")
	}

	hashStr := params[0].(string)
	hashBytes, err := hex.DecodeString(strings.TrimPrefix(hashStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid hash format")
	}
	
	var hash [32]byte
	copy(hash[:], hashBytes)
	
	block := s.blockchain.GetBlockByHash(hash)
	if block == nil {
		return nil, nil
	}

	fullTx := len(params) > 1 && params[1].(bool)
	return s.formatBlock(block, fullTx), nil
}

func (s *Server) ethGetTransactionByHash(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing transaction hash parameter")
	}

	hashStr := params[0].(string)
	hashBytes, err := hex.DecodeString(strings.TrimPrefix(hashStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid hash format")
	}

	var hash [32]byte
	copy(hash[:], hashBytes)
	
	// Check mempool first
	if tx := s.blockchain.GetMempool().GetTransaction(hash); tx != nil {
		return s.formatTransaction(tx, &blockHash{}, 0, 0), nil
	}

	// Search in blocks (this could be optimized with an index)
	currentBlock := s.blockchain.GetCurrentBlock()
	if currentBlock == nil {
		return nil, nil
	}

	for i := uint64(0); i <= currentBlock.Header.Number; i++ {
		block := s.blockchain.GetBlockByNumber(i)
		if block == nil {
			continue
		}

		for txIndex, tx := range block.Transactions {
			if tx.Hash == hash {
				return s.formatTransaction(tx, &block.Header.Hash, block.Header.Number, uint64(txIndex)), nil
			}
		}
	}

	return nil, nil
}

func (s *Server) ethGetTransactionReceipt(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing transaction hash parameter")
	}

	hashStr := params[0].(string)
	hashBytes, err := hex.DecodeString(strings.TrimPrefix(hashStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid hash format")
	}

	var hash [32]byte
	copy(hash[:], hashBytes)

	// Search for transaction receipt in blocks
	currentBlock := s.blockchain.GetCurrentBlock()
	if currentBlock == nil {
		return nil, nil
	}

	for i := uint64(0); i <= currentBlock.Header.Number; i++ {
		block := s.blockchain.GetBlockByNumber(i)
		if block == nil {
			continue
		}

		for txIndex, tx := range block.Transactions {
			if tx.Hash == hash {
				if txIndex < len(block.Receipts) {
					return s.formatReceipt(block.Receipts[txIndex]), nil
				}
			}
		}
	}

	return nil, nil
}

func (s *Server) ethSendRawTransaction(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing raw transaction parameter")
	}

	// In a real implementation, you would decode the raw transaction
	// and add it to the mempool
	return "0x" + strings.Repeat("0", 64), nil
}

func (s *Server) ethCall(params []interface{}) (interface{}, error) {
	// Simulate transaction call without creating a transaction
	return "0x", nil
}

func (s *Server) ethEstimateGas(params []interface{}) (interface{}, error) {
	// Estimate gas for transaction
	return fmt.Sprintf("0x%x", uint64(21000)), nil
}

func (s *Server) ethGasPrice() (interface{}, error) {
	// Return current gas price (20 Gwei)
	return fmt.Sprintf("0x%x", uint64(20000000000)), nil
}

func (s *Server) ethChainId() (interface{}, error) {
	return fmt.Sprintf("0x%x", s.blockchain.GetConfig().ChainID), nil
}

func (s *Server) ethGetTransactionCount(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	addressStr := params[0].(string)
	addressBytes, err := hex.DecodeString(strings.TrimPrefix(addressStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}

	var address [20]byte
	copy(address[:], addressBytes)
	nonce := s.blockchain.GetStateDB().GetNonce(address)
	return fmt.Sprintf("0x%x", nonce), nil
}

func (s *Server) ethGetCode(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	addressStr := params[0].(string)
	addressBytes, err := hex.DecodeString(strings.TrimPrefix(addressStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}

	var address [20]byte
	copy(address[:], addressBytes)
	code := s.blockchain.GetStateDB().GetCode(address)
	return fmt.Sprintf("0x%x", code), nil
}

func (s *Server) ethGetStorageAt(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("missing parameters")
	}

	addressStr := params[0].(string)
	addressBytes, err := hex.DecodeString(strings.TrimPrefix(addressStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid address format")
	}

	var address [20]byte
	copy(address[:], addressBytes)

	keyStr := params[1].(string)
	keyBytes, err := hex.DecodeString(strings.TrimPrefix(keyStr, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid key format")
	}

	var key [32]byte
	copy(key[:], keyBytes)

	value := s.blockchain.GetStateDB().GetState(address, key)
	return fmt.Sprintf("0x%x", value), nil
}

func (s *Server) ethGetLogs(params []interface{}) (interface{}, error) {
	// Return empty logs for now
	return []interface{}{}, nil
}

func (s *Server) netVersion() (interface{}, error) {
	return strconv.FormatUint(s.blockchain.GetConfig().ChainID, 10), nil
}

func (s *Server) formatBlock(block *core.Block, fullTx bool) map[string]interface{} {
	result := map[string]interface{}{
		"number":           fmt.Sprintf("0x%x", block.Header.Number),
		"hash":             fmt.Sprintf("0x%x", block.Header.Hash),
		"parentHash":       fmt.Sprintf("0x%x", block.Header.ParentHash),
		"timestamp":        fmt.Sprintf("0x%x", block.Header.Timestamp),
		"stateRoot":        fmt.Sprintf("0x%x", block.Header.StateRoot),
		"transactionsRoot": fmt.Sprintf("0x%x", block.Header.TxHash),
		"receiptsRoot":     fmt.Sprintf("0x%x", block.Header.ReceiptHash),
		"gasLimit":         fmt.Sprintf("0x%x", block.Header.GasLimit),
		"gasUsed":          fmt.Sprintf("0x%x", block.Header.GasUsed),
		"difficulty":       fmt.Sprintf("0x%x", block.Header.Difficulty),
		"nonce":            fmt.Sprintf("0x%x", block.Header.Nonce),
		"size":             fmt.Sprintf("0x%x", 1000), // Placeholder
	}

	if fullTx {
		var transactions []interface{}
		for i, tx := range block.Transactions {
			transactions = append(transactions, s.formatTransaction(tx, &block.Header.Hash, block.Header.Number, uint64(i)))
		}
		result["transactions"] = transactions
	} else {
		var txHashes []string
		for _, tx := range block.Transactions {
			txHashes = append(txHashes, fmt.Sprintf("0x%x", tx.Hash))
		}
		result["transactions"] = txHashes
	}

	return result
}

func (s *Server) formatTransaction(tx *core.Transaction, blockHash *[32]byte, blockNumber uint64, txIndex uint64) map[string]interface{} {
	result := map[string]interface{}{
		"hash":             fmt.Sprintf("0x%x", tx.Hash),
		"nonce":            fmt.Sprintf("0x%x", tx.Nonce),
		"gasPrice":         fmt.Sprintf("0x%x", tx.GasPrice),
		"gas":              fmt.Sprintf("0x%x", tx.GasLimit),
		"value":            fmt.Sprintf("0x%x", tx.Value),
		"input":            fmt.Sprintf("0x%x", tx.Data),
		"from":             fmt.Sprintf("0x%x", tx.From),
		"v":                fmt.Sprintf("0x%x", tx.V),
		"r":                fmt.Sprintf("0x%x", tx.R),
		"s":                fmt.Sprintf("0x%x", tx.S),
	}

	if tx.To != nil {
		result["to"] = fmt.Sprintf("0x%x", *tx.To)
	} else {
		result["to"] = nil
	}

	if blockHash != nil {
		result["blockHash"] = fmt.Sprintf("0x%x", *blockHash)
		result["blockNumber"] = fmt.Sprintf("0x%x", blockNumber)
		result["transactionIndex"] = fmt.Sprintf("0x%x", txIndex)
	} else {
		result["blockHash"] = nil
		result["blockNumber"] = nil
		result["transactionIndex"] = nil
	}

	return result
}

func (s *Server) formatReceipt(receipt *core.TransactionReceipt) map[string]interface{} {
	return map[string]interface{}{
		"transactionHash":   fmt.Sprintf("0x%x", receipt.TxHash),
		"transactionIndex":  fmt.Sprintf("0x%x", receipt.TxIndex),
		"blockHash":         fmt.Sprintf("0x%x", receipt.BlockHash),
		"blockNumber":       fmt.Sprintf("0x%x", receipt.BlockNumber),
		"from":              fmt.Sprintf("0x%x", receipt.From),
		"to":                func() interface{} {
			if receipt.To != nil {
				return fmt.Sprintf("0x%x", *receipt.To)
			}
			return nil
		}(),
		"gasUsed":           fmt.Sprintf("0x%x", receipt.GasUsed),
		"cumulativeGasUsed": fmt.Sprintf("0x%x", receipt.CumulativeGasUsed),
		"contractAddress":   func() interface{} {
			if receipt.ContractAddress != nil {
				return fmt.Sprintf("0x%x", *receipt.ContractAddress)
			}
			return nil
		}(),
		"logs":              receipt.Logs,
		"status":            fmt.Sprintf("0x%x", receipt.Status),
		"logsBloom":         fmt.Sprintf("0x%x", receipt.LogsBloom),
	}
}

type blockHash [32]byte

func (s *Server) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := JSONRPCResponse{
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: message},
		Version: "2.0",
	}
	json.NewEncoder(w).Encode(response)
}
