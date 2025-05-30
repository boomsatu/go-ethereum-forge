
package core

import (
	"blockchain-node/crypto"
	"encoding/json"
	"math/big"
	"time"
)

type BlockHeader struct {
	Number       uint64      `json:"number"`
	ParentHash   [32]byte    `json:"parentHash"`
	Timestamp    int64       `json:"timestamp"`
	StateRoot    [32]byte    `json:"stateRoot"`
	TxHash       [32]byte    `json:"transactionsRoot"`
	ReceiptHash  [32]byte    `json:"receiptsRoot"`
	LogsBloom    []byte      `json:"logsBloom"`
	GasLimit     uint64      `json:"gasLimit"`
	GasUsed      uint64      `json:"gasUsed"`
	Difficulty   *big.Int    `json:"difficulty"`
	Nonce        uint64      `json:"nonce"`
	Hash         [32]byte    `json:"hash"`
}

type Block struct {
	Header       *BlockHeader           `json:"header"`
	Transactions []*Transaction         `json:"transactions"`
	Receipts     []*TransactionReceipt  `json:"receipts"`
}

func NewBlock(parentHash [32]byte, number uint64, transactions []*Transaction) *Block {
	header := &BlockHeader{
		Number:     number,
		ParentHash: parentHash,
		Timestamp:  time.Now().Unix(),
		GasLimit:   8000000,
		Difficulty: big.NewInt(1000),
	}

	block := &Block{
		Header:       header,
		Transactions: transactions,
		Receipts:     []*TransactionReceipt{},
	}

	return block
}

func (b *Block) CalculateHash() [32]byte {
	// Create hash data from header fields
	data := make([]byte, 0, 256)
	
	// Number (8 bytes)
	numberBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		numberBytes[7-i] = byte(b.Header.Number >> (i * 8))
	}
	data = append(data, numberBytes...)
	
	// Parent hash
	data = append(data, b.Header.ParentHash[:]...)
	
	// Timestamp (8 bytes)
	timestampBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		timestampBytes[7-i] = byte(b.Header.Timestamp >> (i * 8))
	}
	data = append(data, timestampBytes...)
	
	// State root
	data = append(data, b.Header.StateRoot[:]...)
	
	// Transactions root
	data = append(data, b.Header.TxHash[:]...)
	
	// Receipts root
	data = append(data, b.Header.ReceiptHash[:]...)
	
	// Gas limit (8 bytes)
	gasLimitBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		gasLimitBytes[7-i] = byte(b.Header.GasLimit >> (i * 8))
	}
	data = append(data, gasLimitBytes...)
	
	// Gas used (8 bytes)
	gasUsedBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		gasUsedBytes[7-i] = byte(b.Header.GasUsed >> (i * 8))
	}
	data = append(data, gasUsedBytes...)
	
	// Difficulty
	data = append(data, b.Header.Difficulty.Bytes()...)
	
	// Nonce (8 bytes)
	nonceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		nonceBytes[7-i] = byte(b.Header.Nonce >> (i * 8))
	}
	data = append(data, nonceBytes...)
	
	return crypto.SHA256Hash(data)
}

func (b *Block) MineBlock(difficulty *big.Int) {
	target := new(big.Int).Div(crypto.MaxTarget, difficulty)
	
	for {
		b.Header.Nonce++
		hash := b.CalculateHash()
		hashInt := new(big.Int).SetBytes(hash[:])
		
		if hashInt.Cmp(target) == -1 {
			b.Header.Hash = hash
			break
		}
	}
}

func (bh *BlockHeader) ToJSON() ([]byte, error) {
	return json.Marshal(bh)
}

func (b *Block) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}
