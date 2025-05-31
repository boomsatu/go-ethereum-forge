
package rpc

import (
	"blockchain-node/core"
	"blockchain-node/crypto"
	"blockchain-node/wallet"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
)

type WalletAPI struct {
	blockchain *core.Blockchain
}

func NewWalletAPI(blockchain *core.Blockchain) *WalletAPI {
	return &WalletAPI{
		blockchain: blockchain,
	}
}

func (api *WalletAPI) CreateHandler(w http.ResponseWriter, r *http.Request) {
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

	newWallet, err := wallet.NewWallet()
	if err != nil {
		http.Error(w, "Failed to create wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get balance (should be 0 for new wallet)
	address := newWallet.GetAddressBytes()
	balance := api.blockchain.GetStateDB().GetBalance(address)

	// Format private key with 0x prefix and ensure 64 characters
	privateKeyHex := newWallet.GetPrivateKeyHex()
	if !strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = "0x" + privateKeyHex
	}
	// Ensure 64 character hex string (32 bytes)
	if len(privateKeyHex) == 66 { // 0x + 64 chars
		// Good
	} else if len(privateKeyHex) < 66 {
		// Pad with zeros
		privateKeyHex = "0x" + strings.Repeat("0", 66-len(privateKeyHex)) + privateKeyHex[2:]
	}

	response := map[string]interface{}{
		"address":    "0x" + newWallet.GetAddress(),
		"privateKey": privateKeyHex,
		"publicKey":  "0x" + newWallet.GetPublicKeyHex(),
		"balance":    fmt.Sprintf("0x%x", balance),
		"balanceEth": formatWeiToEth(balance),
	}

	json.NewEncoder(w).Encode(response)
}

func (api *WalletAPI) ImportHandler(w http.ResponseWriter, r *http.Request) {
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
		PrivateKey string `json:"privateKey"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Clean private key
	privateKeyHex := strings.TrimSpace(req.PrivateKey)
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	// Validate private key length
	if len(privateKeyHex) != 64 {
		http.Error(w, "Private key must be 64 hex characters", http.StatusBadRequest)
		return
	}

	importedWallet, err := wallet.NewWalletFromPrivateKey(privateKeyHex)
	if err != nil {
		http.Error(w, "Failed to import wallet: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get balance
	address := importedWallet.GetAddressBytes()
	balance := api.blockchain.GetStateDB().GetBalance(address)

	response := map[string]interface{}{
		"address":    "0x" + importedWallet.GetAddress(),
		"privateKey": "0x" + importedWallet.GetPrivateKeyHex(),
		"publicKey":  "0x" + importedWallet.GetPublicKeyHex(),
		"balance":    fmt.Sprintf("0x%x", balance),
		"balanceEth": formatWeiToEth(balance),
		"valid":      true,
	}

	json.NewEncoder(w).Encode(response)
}

func (api *WalletAPI) SendTransactionHandler(w http.ResponseWriter, r *http.Request) {
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
		From       string `json:"from"`
		To         string `json:"to"`
		Value      string `json:"value"`
		GasLimit   string `json:"gasLimit"`
		GasPrice   string `json:"gasPrice"`
		PrivateKey string `json:"privateKey"`
		Data       string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Import wallet from private key
	privateKeyHex := strings.TrimSpace(req.PrivateKey)
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	senderWallet, err := wallet.NewWalletFromPrivateKey(privateKeyHex)
	if err != nil {
		http.Error(w, "Invalid private key", http.StatusBadRequest)
		return
	}

	// Verify sender address matches
	if strings.ToLower("0x"+senderWallet.GetAddress()) != strings.ToLower(req.From) {
		http.Error(w, "Private key does not match sender address", http.StatusBadRequest)
		return
	}

	// Parse values
	value, ok := new(big.Int).SetString(req.Value, 0)
	if !ok {
		http.Error(w, "Invalid value format", http.StatusBadRequest)
		return
	}

	gasLimit, ok := new(big.Int).SetString(req.GasLimit, 0)
	if !ok {
		gasLimit = big.NewInt(21000) // Default gas limit
	}

	gasPrice, ok := new(big.Int).SetString(req.GasPrice, 0)
	if !ok {
		gasPrice = big.NewInt(20000000000) // Default 20 Gwei
	}

	// Parse to address
	var toAddr *[20]byte
	if req.To != "" {
		toAddrStr := strings.TrimSpace(req.To)
		if strings.HasPrefix(toAddrStr, "0x") {
			toAddrStr = toAddrStr[2:]
		}
		
		if len(toAddrStr) == 40 {
			toBytes := crypto.HexToBytes(toAddrStr)
			if len(toBytes) == 20 {
				var addr [20]byte
				copy(addr[:], toBytes)
				toAddr = &addr
			}
		}
		
		if toAddr == nil {
			http.Error(w, "Invalid to address format", http.StatusBadRequest)
			return
		}
	}

	// Get nonce
	fromAddr := senderWallet.GetAddressBytes()
	nonce := api.blockchain.GetStateDB().GetNonce(fromAddr)

	// Parse data
	var data []byte
	if req.Data != "" {
		dataStr := strings.TrimSpace(req.Data)
		if strings.HasPrefix(dataStr, "0x") {
			dataStr = dataStr[2:]
		}
		data = crypto.HexToBytes(dataStr)
	}

	// Check balance
	balance := api.blockchain.GetStateDB().GetBalance(fromAddr)
	totalCost := new(big.Int).Add(value, new(big.Int).Mul(gasPrice, gasLimit))
	if balance.Cmp(totalCost) < 0 {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Create transaction
	tx := core.NewTransaction(nonce, toAddr, value, gasLimit.Uint64(), gasPrice, data)

	// Sign transaction
	if err := senderWallet.SignTransaction(tx); err != nil {
		http.Error(w, "Failed to sign transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add to mempool
	if err := api.blockchain.GetMempool().AddTransaction(tx); err != nil {
		http.Error(w, "Failed to add transaction to mempool: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"hash":    fmt.Sprintf("0x%x", tx.Hash),
		"success": true,
		"nonce":   fmt.Sprintf("0x%x", nonce),
		"gasUsed": fmt.Sprintf("0x%x", gasLimit.Uint64()),
	}

	json.NewEncoder(w).Encode(response)
}

// CheckBalanceHandler checks balance for an address
func (api *WalletAPI) CheckBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter is required", http.StatusBadRequest)
		return
	}

	// Clean address
	address = strings.TrimSpace(address)
	if strings.HasPrefix(address, "0x") {
		address = address[2:]
	}

	if len(address) != 40 {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	addrBytes := crypto.HexToBytes(address)
	if len(addrBytes) != 20 {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	var addr [20]byte
	copy(addr[:], addrBytes)

	balance := api.blockchain.GetStateDB().GetBalance(addr)
	nonce := api.blockchain.GetStateDB().GetNonce(addr)

	response := map[string]interface{}{
		"address":    "0x" + address,
		"balance":    fmt.Sprintf("0x%x", balance),
		"balanceEth": formatWeiToEth(balance),
		"nonce":      fmt.Sprintf("0x%x", nonce),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function to format Wei to ETH
func formatWeiToEth(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	
	// Convert wei to ETH (1 ETH = 10^18 wei)
	eth := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	eth.Quo(eth, divisor)
	
	return eth.Text('f', 6)
}

// Helper function to parse hex string to bytes
func parseHexToBytes(hexStr string, expectedLen int) ([]byte, error) {
	hexStr = strings.TrimSpace(hexStr)
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}
	
	if len(hexStr) != expectedLen*2 {
		return nil, fmt.Errorf("expected %d bytes (got %d)", expectedLen, len(hexStr)/2)
	}
	
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %v", err)
	}
	
	return bytes, nil
}
