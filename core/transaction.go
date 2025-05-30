
package core

import (
	"blockchain-node/crypto"
	"encoding/hex"
	"encoding/json"
	"math/big"
)

type Transaction struct {
	Nonce    uint64      `json:"nonce"`
	GasPrice *big.Int    `json:"gasPrice"`
	GasLimit uint64      `json:"gasLimit"`
	To       *[20]byte   `json:"to"`
	Value    *big.Int    `json:"value"`
	Data     []byte      `json:"data"`
	V        *big.Int    `json:"v"`
	R        *big.Int    `json:"r"`
	S        *big.Int    `json:"s"`
	Hash     [32]byte    `json:"hash"`
	From     [20]byte    `json:"from"`
}

type TransactionReceipt struct {
	TxHash          [32]byte       `json:"transactionHash"`
	TxIndex         uint64         `json:"transactionIndex"`
	BlockHash       [32]byte       `json:"blockHash"`
	BlockNumber     uint64         `json:"blockNumber"`
	From            [20]byte       `json:"from"`
	To              *[20]byte      `json:"to"`
	GasUsed         uint64         `json:"gasUsed"`
	CumulativeGasUsed uint64       `json:"cumulativeGasUsed"`
	ContractAddress *[20]byte      `json:"contractAddress"`
	Logs            []*Log         `json:"logs"`
	Status          uint64         `json:"status"`
	LogsBloom       []byte         `json:"logsBloom"`
}

type Log struct {
	Address     [20]byte   `json:"address"`
	Topics      [][32]byte `json:"topics"`
	Data        []byte     `json:"data"`
	BlockNumber uint64     `json:"blockNumber"`
	TxHash      [32]byte   `json:"transactionHash"`
	TxIndex     uint64     `json:"transactionIndex"`
	BlockHash   [32]byte   `json:"blockHash"`
	Index       uint64     `json:"logIndex"`
	Removed     bool       `json:"removed"`
}

func NewTransaction(nonce uint64, to *[20]byte, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	tx := &Transaction{
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		To:       to,
		Value:    value,
		Data:     data,
	}
	
	tx.Hash = tx.CalculateHash()
	return tx
}

func (tx *Transaction) CalculateHash() [32]byte {
	data, _ := json.Marshal(struct {
		Nonce    uint64 `json:"nonce"`
		GasPrice string `json:"gasPrice"`
		GasLimit uint64 `json:"gasLimit"`
		To       string `json:"to"`
		Value    string `json:"value"`
		Data     string `json:"data"`
	}{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice.String(),
		GasLimit: tx.GasLimit,
		To: func() string {
			if tx.To != nil {
				return hex.EncodeToString(tx.To[:])
			}
			return ""
		}(),
		Value: tx.Value.String(),
		Data:  hex.EncodeToString(tx.Data),
	})
	
	return crypto.Keccak256Hash(data)
}

func (tx *Transaction) Sign(privateKey []byte) error {
	hash := tx.CalculateHash()
	
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return err
	}
	
	// Extract V, R, S from signature
	tx.V = big.NewInt(int64(signature[64]) + 27)
	tx.R = new(big.Int).SetBytes(signature[:32])
	tx.S = new(big.Int).SetBytes(signature[32:64])
	
	// Recover sender address
	recoveredAddr, err := crypto.RecoverAddress(hash[:], signature)
	if err != nil {
		return err
	}
	
	copy(tx.From[:], recoveredAddr[:])
	return nil
}

func (tx *Transaction) VerifySignature() bool {
	hash := tx.CalculateHash()
	
	// Reconstruct signature
	signature := make([]byte, 65)
	copy(signature[:32], tx.R.Bytes())
	copy(signature[32:64], tx.S.Bytes())
	signature[64] = byte(tx.V.Int64() - 27)
	
	recoveredAddr, err := crypto.RecoverAddress(hash[:], signature)
	if err != nil {
		return false
	}
	
	return recoveredAddr == tx.From
}

func (tx *Transaction) IsContractCreation() bool {
	return tx.To == nil
}

func (tx *Transaction) ToJSON() ([]byte, error) {
	return json.Marshal(tx)
}
