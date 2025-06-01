
package core

import (
	"blockchain-node/crypto"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Transaction struct {
	Nonce    uint64           `json:"nonce"`
	To       *common.Address  `json:"to"`
	Value    *big.Int         `json:"value"`
	GasLimit uint64           `json:"gasLimit"`
	GasPrice *big.Int         `json:"gasPrice"`
	Data     []byte           `json:"data"`
	V        *big.Int         `json:"v"`
	R        *big.Int         `json:"r"`
	S        *big.Int         `json:"s"`
	Hash     [32]byte         `json:"hash"`
	From     common.Address   `json:"from"`
}

// Implement validation interfaces
func (tx *Transaction) GetHash() [32]byte { return tx.Hash }
func (tx *Transaction) GetFrom() common.Address { return tx.From }
func (tx *Transaction) GetTo() *common.Address { return tx.To }
func (tx *Transaction) GetValue() *big.Int { return tx.Value }
func (tx *Transaction) GetGasPrice() *big.Int { return tx.GasPrice }
func (tx *Transaction) GetGasLimit() uint64 { return tx.GasLimit }
func (tx *Transaction) GetData() []byte { return tx.Data }
func (tx *Transaction) GetV() *big.Int { return tx.V }
func (tx *Transaction) GetR() *big.Int { return tx.R }
func (tx *Transaction) GetS() *big.Int { return tx.S }

func NewTransaction(nonce uint64, to *common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	tx := &Transaction{
		Nonce:    nonce,
		To:       to,
		Value:    value,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	}
	
	tx.Hash = tx.CalculateHash()
	return tx
}

func (tx *Transaction) CalculateHash() [32]byte {
	data := make([]byte, 0, 256)
	
	// Nonce (8 bytes)
	nonceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		nonceBytes[7-i] = byte(tx.Nonce >> (i * 8))
	}
	data = append(data, nonceBytes...)
	
	// To address (20 bytes, or empty if nil)
	if tx.To != nil {
		data = append(data, tx.To.Bytes()...)
	} else {
		data = append(data, make([]byte, 20)...)
	}
	
	// Value
	if tx.Value != nil {
		data = append(data, tx.Value.Bytes()...)
	}
	
	// Gas limit (8 bytes)
	gasLimitBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		gasLimitBytes[7-i] = byte(tx.GasLimit >> (i * 8))
	}
	data = append(data, gasLimitBytes...)
	
	// Gas price
	if tx.GasPrice != nil {
		data = append(data, tx.GasPrice.Bytes()...)
	}
	
	// Data
	data = append(data, tx.Data...)
	
	return crypto.SHA256Hash(data)
}

func (tx *Transaction) VerifySignature() bool {
	// Simplified signature verification
	// In a real implementation, this would verify the ECDSA signature
	return tx.V != nil && tx.R != nil && tx.S != nil
}

func (tx *Transaction) ToJSON() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *Transaction) FromJSON(data []byte) error {
	return json.Unmarshal(data, tx)
}

func (tx *Transaction) IsContractCreation() bool {
	return tx.To == nil
}

func (tx *Transaction) ToEthTransaction() *ethTypes.Transaction {
	var to *common.Address
	if tx.To != nil {
		to = tx.To
	}
	
	ethTx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    tx.Nonce,
		To:       to,
		Value:    tx.Value,
		Gas:      tx.GasLimit,
		GasPrice: tx.GasPrice,
		Data:     tx.Data,
		V:        tx.V,
		R:        tx.R,
		S:        tx.S,
	})
	
	return ethTx
}

type TransactionReceipt struct {
	TxHash            [32]byte        `json:"transactionHash"`
	TxIndex           uint64          `json:"transactionIndex"`
	BlockHash         [32]byte        `json:"blockHash"`
	BlockNumber       uint64          `json:"blockNumber"`
	From              common.Address  `json:"from"`
	To                *common.Address `json:"to"`
	ContractAddress   *common.Address `json:"contractAddress"`
	GasUsed           uint64          `json:"gasUsed"`
	CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
	Status            uint64          `json:"status"`
	Logs              []*Log          `json:"logs"`
}

type Log struct {
	Address     common.Address   `json:"address"`
	Topics      []common.Hash    `json:"topics"`
	Data        []byte           `json:"data"`
	BlockNumber uint64           `json:"blockNumber"`
	TxHash      [32]byte         `json:"transactionHash"`
	TxIndex     uint64           `json:"transactionIndex"`
	BlockHash   [32]byte         `json:"blockHash"`
	Index       uint64           `json:"logIndex"`
	Removed     bool             `json:"removed"`
}
