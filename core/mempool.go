
package core

import (
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Mempool struct {
	transactions map[common.Hash]*Transaction
	pending      map[common.Address][]*Transaction
	mu           sync.RWMutex
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[common.Hash]*Transaction),
		pending:      make(map[common.Address][]*Transaction),
	}
}

func (mp *Mempool) AddTransaction(tx *Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Validate transaction
	if err := mp.validateTransaction(tx); err != nil {
		return err
	}

	// Add to mempool
	mp.transactions[tx.Hash] = tx
	mp.pending[tx.From] = append(mp.pending[tx.From], tx)

	return nil
}

func (mp *Mempool) validateTransaction(tx *Transaction) error {
	// Check if transaction already exists
	if _, exists := mp.transactions[tx.Hash]; exists {
		return errors.New("transaction already exists in mempool")
	}

	// Verify signature
	if !tx.VerifySignature() {
		return errors.New("invalid transaction signature")
	}

	// Additional validations can be added here
	// - Check nonce
	// - Check gas price
	// - Check balance

	return nil
}

func (mp *Mempool) GetTransaction(hash common.Hash) *Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.transactions[hash]
}

func (mp *Mempool) GetPendingTransactions() []*Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var txs []*Transaction
	for _, tx := range mp.transactions {
		txs = append(txs, tx)
	}
	return txs
}

func (mp *Mempool) RemoveTransaction(hash common.Hash) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if tx, exists := mp.transactions[hash]; exists {
		delete(mp.transactions, hash)
		
		// Remove from pending
		if pending := mp.pending[tx.From]; pending != nil {
			for i, pendingTx := range pending {
				if pendingTx.Hash == hash {
					mp.pending[tx.From] = append(pending[:i], pending[i+1:]...)
					break
				}
			}
		}
	}
}

func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions = make(map[common.Hash]*Transaction)
	mp.pending = make(map[common.Address][]*Transaction)
}

func (mp *Mempool) Size() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.transactions)
}
