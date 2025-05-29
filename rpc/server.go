
package rpc

import (
	"blockchain-node/core"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/mux"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	config     *Config
	blockchain *core.Blockchain
	server     *http.Server
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
	return &Server{
		config:     config,
		blockchain: blockchain,
	}
}

func (s *Server) Start() error {
	router := mux.NewRouter()
	router.HandleFunc("/", s.handleRPC).Methods("POST")
	router.HandleFunc("/health", s.handleHealth).Methods("GET")

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

	fmt.Printf("JSON-RPC server started on %s\n", addr)
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
	return hexutil.EncodeUint64(currentBlock.Header.Number), nil
}

func (s *Server) ethGetBalance(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	address := common.HexToAddress(params[0].(string))
	balance := s.blockchain.GetStateDB().GetBalance(address)
	
	return hexutil.EncodeBig(balance), nil
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
		blockNum, err = hexutil.DecodeUint64(blockNumStr)
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

	hash := common.HexToHash(params[0].(string))
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

	hash := common.HexToHash(params[0].(string))
	
	// Check mempool first
	if tx := s.blockchain.GetMempool().GetTransaction(hash); tx != nil {
		return s.formatTransaction(tx, nil, 0, 0), nil
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

	hash := common.HexToHash(params[0].(string))

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
	return hexutil.EncodeUint64(21000), nil
}

func (s *Server) ethGasPrice() (interface{}, error) {
	// Return current gas price (20 Gwei)
	return hexutil.EncodeUint64(20000000000), nil
}

func (s *Server) ethChainId() (interface{}, error) {
	return hexutil.EncodeUint64(s.blockchain.GetConfig().ChainID), nil
}

func (s *Server) ethGetTransactionCount(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	address := common.HexToAddress(params[0].(string))
	nonce := s.blockchain.GetStateDB().GetNonce(address)
	
	return hexutil.EncodeUint64(nonce), nil
}

func (s *Server) ethGetCode(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	address := common.HexToAddress(params[0].(string))
	code := s.blockchain.GetStateDB().GetCode(address)
	
	return hexutil.Encode(code), nil
}

func (s *Server) ethGetStorageAt(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("missing parameters")
	}

	address := common.HexToAddress(params[0].(string))
	key := common.HexToHash(params[1].(string))
	value := s.blockchain.GetStateDB().GetState(address, key)
	
	return value.Hex(), nil
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
		"number":           hexutil.EncodeUint64(block.Header.Number),
		"hash":             block.Header.Hash.Hex(),
		"parentHash":       block.Header.ParentHash.Hex(),
		"timestamp":        hexutil.EncodeUint64(uint64(block.Header.Timestamp)),
		"stateRoot":        block.Header.StateRoot.Hex(),
		"transactionsRoot": block.Header.TxHash.Hex(),
		"receiptsRoot":     block.Header.ReceiptHash.Hex(),
		"gasLimit":         hexutil.EncodeUint64(block.Header.GasLimit),
		"gasUsed":          hexutil.EncodeUint64(block.Header.GasUsed),
		"difficulty":       hexutil.EncodeBig(block.Header.Difficulty),
		"nonce":            hexutil.EncodeUint64(block.Header.Nonce),
		"size":             hexutil.EncodeUint64(1000), // Placeholder
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
			txHashes = append(txHashes, tx.Hash.Hex())
		}
		result["transactions"] = txHashes
	}

	return result
}

func (s *Server) formatTransaction(tx *core.Transaction, blockHash *common.Hash, blockNumber uint64, txIndex uint64) map[string]interface{} {
	result := map[string]interface{}{
		"hash":             tx.Hash.Hex(),
		"nonce":            hexutil.EncodeUint64(tx.Nonce),
		"gasPrice":         hexutil.EncodeBig(tx.GasPrice),
		"gas":              hexutil.EncodeUint64(tx.GasLimit),
		"value":            hexutil.EncodeBig(tx.Value),
		"input":            hexutil.Encode(tx.Data),
		"from":             tx.From.Hex(),
		"v":                hexutil.EncodeBig(tx.V),
		"r":                hexutil.EncodeBig(tx.R),
		"s":                hexutil.EncodeBig(tx.S),
	}

	if tx.To != nil {
		result["to"] = tx.To.Hex()
	} else {
		result["to"] = nil
	}

	if blockHash != nil {
		result["blockHash"] = blockHash.Hex()
		result["blockNumber"] = hexutil.EncodeUint64(blockNumber)
		result["transactionIndex"] = hexutil.EncodeUint64(txIndex)
	} else {
		result["blockHash"] = nil
		result["blockNumber"] = nil
		result["transactionIndex"] = nil
	}

	return result
}

func (s *Server) formatReceipt(receipt *core.TransactionReceipt) map[string]interface{} {
	return map[string]interface{}{
		"transactionHash":   receipt.TxHash.Hex(),
		"transactionIndex":  hexutil.EncodeUint64(receipt.TxIndex),
		"blockHash":         receipt.BlockHash.Hex(),
		"blockNumber":       hexutil.EncodeUint64(receipt.BlockNumber),
		"from":              receipt.From.Hex(),
		"to":                func() interface{} {
			if receipt.To != nil {
				return receipt.To.Hex()
			}
			return nil
		}(),
		"gasUsed":           hexutil.EncodeUint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.EncodeUint64(receipt.CumulativeGasUsed),
		"contractAddress":   func() interface{} {
			if receipt.ContractAddress != nil {
				return receipt.ContractAddress.Hex()
			}
			return nil
		}(),
		"logs":              receipt.Logs,
		"status":            hexutil.EncodeUint64(receipt.Status),
		"logsBloom":         hexutil.Encode(receipt.LogsBloom),
	}
}

func (s *Server) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := JSONRPCResponse{
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: message},
		Version: "2.0",
	}
	json.NewEncoder(w).Encode(response)
}
