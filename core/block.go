
package core

import (
	"blockchain-node/crypto"
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type BlockHeader struct {
	Number       uint64      `json:"number"`
	ParentHash   common.Hash `json:"parentHash"`
	Timestamp    int64       `json:"timestamp"`
	StateRoot    common.Hash `json:"stateRoot"`
	TxHash       common.Hash `json:"transactionsRoot"`
	ReceiptHash  common.Hash `json:"receiptsRoot"`
	LogsBloom    []byte      `json:"logsBloom"`
	GasLimit     uint64      `json:"gasLimit"`
	GasUsed      uint64      `json:"gasUsed"`
	Difficulty   *big.Int    `json:"difficulty"`
	Nonce        uint64      `json:"nonce"`
	Hash         common.Hash `json:"hash"`
}

type Block struct {
	Header       *BlockHeader           `json:"header"`
	Transactions []*Transaction         `json:"transactions"`
	Receipts     []*TransactionReceipt  `json:"receipts"`
}

func NewBlock(parentHash common.Hash, number uint64, transactions []*Transaction) *Block {
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

func (b *Block) CalculateHash() common.Hash {
	// Serialize block header for hashing
	data, _ := json.Marshal(b.Header)
	return crypto.Keccak256Hash(data)
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
