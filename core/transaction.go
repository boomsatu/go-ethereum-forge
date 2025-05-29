
package core

import (
	"blockchain-node/crypto"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	Nonce    uint64      `json:"nonce"`
	GasPrice *big.Int    `json:"gasPrice"`
	GasLimit uint64      `json:"gasLimit"`
	To       *common.Address `json:"to"`
	Value    *big.Int    `json:"value"`
	Data     []byte      `json:"data"`
	V        *big.Int    `json:"v"`
	R        *big.Int    `json:"r"`
	S        *big.Int    `json:"s"`
	Hash     common.Hash `json:"hash"`
	From     common.Address `json:"from"`
}

type TransactionReceipt struct {
	TxHash          common.Hash    `json:"transactionHash"`
	TxIndex         uint64         `json:"transactionIndex"`
	BlockHash       common.Hash    `json:"blockHash"`
	BlockNumber     uint64         `json:"blockNumber"`
	From            common.Address `json:"from"`
	To              *common.Address `json:"to"`
	GasUsed         uint64         `json:"gasUsed"`
	CumulativeGasUsed uint64       `json:"cumulativeGasUsed"`
	ContractAddress *common.Address `json:"contractAddress"`
	Logs            []*Log         `json:"logs"`
	Status          uint64         `json:"status"`
	LogsBloom       []byte         `json:"logsBloom"`
}

type Log struct {
	Address     common.Address `json:"address"`
	Topics      []common.Hash  `json:"topics"`
	Data        []byte         `json:"data"`
	BlockNumber uint64         `json:"blockNumber"`
	TxHash      common.Hash    `json:"transactionHash"`
	TxIndex     uint64         `json:"transactionIndex"`
	BlockHash   common.Hash    `json:"blockHash"`
	Index       uint64         `json:"logIndex"`
	Removed     bool           `json:"removed"`
}

func NewTransaction(nonce uint64, to *common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
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

func (tx *Transaction) CalculateHash() common.Hash {
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
		To:       func() string {
			if tx.To != nil {
				return tx.To.Hex()
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
	pubKey, err := crypto.Ecrecover(hash[:], signature)
	if err != nil {
		return err
	}
	
	tx.From = crypto.PubkeyToAddress(pubKey)
	return nil
}

func (tx *Transaction) VerifySignature() bool {
	hash := tx.CalculateHash()
	
	// Reconstruct signature
	signature := make([]byte, 65)
	copy(signature[:32], tx.R.Bytes())
	copy(signature[32:64], tx.S.Bytes())
	signature[64] = byte(tx.V.Int64() - 27)
	
	pubKey, err := crypto.Ecrecover(hash[:], signature)
	if err != nil {
		return false
	}
	
	recoveredAddr := crypto.PubkeyToAddress(pubKey)
	return recoveredAddr == tx.From
}

func (tx *Transaction) ToEthTransaction() *ethTypes.Transaction {
	var to *common.Address
	if tx.To != nil {
		to = tx.To
	}
	
	return ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		Gas:      tx.GasLimit,
		To:       to,
		Value:    tx.Value,
		Data:     tx.Data,
		V:        tx.V,
		R:        tx.R,
		S:        tx.S,
	})
}

func (tx *Transaction) IsContractCreation() bool {
	return tx.To == nil
}

func (tx *Transaction) ToJSON() ([]byte, error) {
	return json.Marshal(tx)
}
