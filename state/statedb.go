
package state

import (
	"blockchain-node/crypto"
	"blockchain-node/database"
	"blockchain-node/trie"
	"encoding/json"
	"fmt"
	"math/big"
)

// Account represents an account in the state
type Account struct {
	Nonce    uint64      `json:"nonce"`
	Balance  *big.Int    `json:"balance"`
	CodeHash [32]byte    `json:"codeHash"`
	Root     [32]byte    `json:"storageRoot"` // Storage trie root
}

// StateDB manages the world state
type StateDB struct {
	db          database.Database
	trie        *trie.Trie
	accounts    map[[20]byte]*Account
	codes       map[[32]byte][]byte
	storage     map[[20]byte]map[[32]byte][32]byte
	logs        []*Log
	snapshots   []*StateSnapshot
	dirty       map[[20]byte]bool
}

// Log represents a log entry
type Log struct {
	Address [20]byte   `json:"address"`
	Topics  [][32]byte `json:"topics"`
	Data    []byte     `json:"data"`
}

// StateSnapshot represents a snapshot of the state
type StateSnapshot struct {
	accounts map[[20]byte]*Account
	codes    map[[32]byte][]byte
	storage  map[[20]byte]map[[32]byte][32]byte
}

// NewStateDB creates a new state database
func NewStateDB(root [32]byte, db database.Database) (*StateDB, error) {
	stateTrie, err := trie.NewTrie(root, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create state trie: %v", err)
	}
	
	return &StateDB{
		db:       db,
		trie:     stateTrie,
		accounts: make(map[[20]byte]*Account),
		codes:    make(map[[32]byte][]byte),
		storage:  make(map[[20]byte]map[[32]byte][32]byte),
		logs:     make([]*Log, 0),
		dirty:    make(map[[20]byte]bool),
	}, nil
}

// GetAccount retrieves an account from the state
func (s *StateDB) GetAccount(addr [20]byte) *Account {
	// Check cache first
	if acc, exists := s.accounts[addr]; exists {
		return acc
	}
	
	// Load from trie
	data, err := s.trie.Get(addr[:])
	if err != nil || data == nil {
		// Account doesn't exist, return empty account
		acc := &Account{
			Nonce:   0,
			Balance: big.NewInt(0),
		}
		s.accounts[addr] = acc
		return acc
	}
	
	var acc Account
	if err := json.Unmarshal(data, &acc); err != nil {
		// Invalid data, return empty account
		acc := &Account{
			Nonce:   0,
			Balance: big.NewInt(0),
		}
		s.accounts[addr] = &acc
		return &acc
	}
	
	s.accounts[addr] = &acc
	return &acc
}

// SetAccount sets an account in the state
func (s *StateDB) SetAccount(addr [20]byte, acc *Account) {
	s.accounts[addr] = acc
	s.dirty[addr] = true
}

// GetBalance gets the balance of an account
func (s *StateDB) GetBalance(addr [20]byte) *big.Int {
	acc := s.GetAccount(addr)
	return new(big.Int).Set(acc.Balance)
}

// SetBalance sets the balance of an account
func (s *StateDB) SetBalance(addr [20]byte, balance *big.Int) {
	acc := s.GetAccount(addr)
	acc.Balance = new(big.Int).Set(balance)
	s.SetAccount(addr, acc)
}

// GetNonce gets the nonce of an account
func (s *StateDB) GetNonce(addr [20]byte) uint64 {
	acc := s.GetAccount(addr)
	return acc.Nonce
}

// SetNonce sets the nonce of an account
func (s *StateDB) SetNonce(addr [20]byte, nonce uint64) {
	acc := s.GetAccount(addr)
	acc.Nonce = nonce
	s.SetAccount(addr, acc)
}

// GetCode gets the code of an account
func (s *StateDB) GetCode(addr [20]byte) []byte {
	acc := s.GetAccount(addr)
	
	// Empty code hash means no code
	emptyHash := [32]byte{}
	if acc.CodeHash == emptyHash {
		return nil
	}
	
	// Check cache first
	if code, exists := s.codes[acc.CodeHash]; exists {
		return code
	}
	
	// Load from database
	key := append([]byte("code_"), acc.CodeHash[:]...)
	data, err := s.db.Get(key)
	if err != nil {
		return nil
	}
	
	s.codes[acc.CodeHash] = data
	return data
}

// SetCode sets the code of an account
func (s *StateDB) SetCode(addr [20]byte, code []byte) {
	acc := s.GetAccount(addr)
	
	if len(code) == 0 {
		acc.CodeHash = [32]byte{}
	} else {
		hash := crypto.Keccak256Hash(code)
		copy(acc.CodeHash[:], hash[:])
		
		// Store code in database
		key := append([]byte("code_"), acc.CodeHash[:]...)
		s.db.Put(key, code)
		
		// Cache code
		s.codes[acc.CodeHash] = code
	}
	
	s.SetAccount(addr, acc)
}

// GetState gets a storage value
func (s *StateDB) GetState(addr [20]byte, key [32]byte) [32]byte {
	// Check cache first
	if storage, exists := s.storage[addr]; exists {
		if value, exists := storage[key]; exists {
			return value
		}
	}
	
	// Load from storage trie
	acc := s.GetAccount(addr)
	if acc.Root == ([32]byte{}) {
		return [32]byte{} // Empty storage
	}
	
	storageTrie, err := trie.NewTrie(acc.Root, s.db)
	if err != nil {
		return [32]byte{}
	}
	
	data, err := storageTrie.Get(key[:])
	if err != nil || data == nil {
		return [32]byte{}
	}
	
	var value [32]byte
	copy(value[:], data)
	
	// Cache the value
	if s.storage[addr] == nil {
		s.storage[addr] = make(map[[32]byte][32]byte)
	}
	s.storage[addr][key] = value
	
	return value
}

// SetState sets a storage value
func (s *StateDB) SetState(addr [20]byte, key [32]byte, value [32]byte) {
	// Update cache
	if s.storage[addr] == nil {
		s.storage[addr] = make(map[[32]byte][32]byte)
	}
	s.storage[addr][key] = value
	s.dirty[addr] = true
}

// AddLog adds a log entry
func (s *StateDB) AddLog(log *Log) {
	s.logs = append(s.logs, log)
}

// GetLogs returns all logs
func (s *StateDB) GetLogs() []*Log {
	return s.logs
}

// Snapshot creates a snapshot of the current state
func (s *StateDB) Snapshot() int {
	snap := &StateSnapshot{
		accounts: make(map[[20]byte]*Account),
		codes:    make(map[[32]byte][]byte),
		storage:  make(map[[20]byte]map[[32]byte][32]byte),
	}
	
	// Copy accounts
	for addr, acc := range s.accounts {
		snap.accounts[addr] = &Account{
			Nonce:    acc.Nonce,
			Balance:  new(big.Int).Set(acc.Balance),
			CodeHash: acc.CodeHash,
			Root:     acc.Root,
		}
	}
	
	// Copy codes
	for hash, code := range s.codes {
		snap.codes[hash] = make([]byte, len(code))
		copy(snap.codes[hash], code)
	}
	
	// Copy storage
	for addr, storage := range s.storage {
		snap.storage[addr] = make(map[[32]byte][32]byte)
		for key, value := range storage {
			snap.storage[addr][key] = value
		}
	}
	
	s.snapshots = append(s.snapshots, snap)
	return len(s.snapshots) - 1
}

// RevertToSnapshot reverts state to a snapshot
func (s *StateDB) RevertToSnapshot(snapId int) {
	if snapId < 0 || snapId >= len(s.snapshots) {
		return
	}
	
	snap := s.snapshots[snapId]
	
	// Restore accounts
	s.accounts = make(map[[20]byte]*Account)
	for addr, acc := range snap.accounts {
		s.accounts[addr] = &Account{
			Nonce:    acc.Nonce,
			Balance:  new(big.Int).Set(acc.Balance),
			CodeHash: acc.CodeHash,
			Root:     acc.Root,
		}
	}
	
	// Restore codes
	s.codes = make(map[[32]byte][]byte)
	for hash, code := range snap.codes {
		s.codes[hash] = make([]byte, len(code))
		copy(s.codes[hash], code)
	}
	
	// Restore storage
	s.storage = make(map[[20]byte]map[[32]byte][32]byte)
	for addr, storage := range snap.storage {
		s.storage[addr] = make(map[[32]byte][32]byte)
		for key, value := range storage {
			s.storage[addr][key] = value
		}
	}
	
	// Remove snapshots after the reverted one
	s.snapshots = s.snapshots[:snapId]
	
	// Reset dirty flags
	s.dirty = make(map[[20]byte]bool)
}

// Commit commits the state changes to the trie
func (s *StateDB) Commit() ([32]byte, error) {
	// Update storage tries for dirty accounts
	for addr := range s.dirty {
		if err := s.updateStorageTrie(addr); err != nil {
			return [32]byte{}, fmt.Errorf("failed to update storage trie for %x: %v", addr, err)
		}
	}
	
	// Update account data in state trie
	for addr, acc := range s.accounts {
		if s.dirty[addr] {
			data, err := json.Marshal(acc)
			if err != nil {
				return [32]byte{}, fmt.Errorf("failed to marshal account %x: %v", addr, err)
			}
			
			if err := s.trie.Update(addr[:], data); err != nil {
				return [32]byte{}, fmt.Errorf("failed to update account %x in trie: %v", addr, err)
			}
		}
	}
	
	// Commit trie changes
	root, err := s.trie.Commit()
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to commit state trie: %v", err)
	}
	
	// Clear dirty flags
	s.dirty = make(map[[20]byte]bool)
	
	// Clear logs
	s.logs = make([]*Log, 0)
	
	return root, nil
}

// updateStorageTrie updates the storage trie for an account
func (s *StateDB) updateStorageTrie(addr [20]byte) error {
	acc := s.GetAccount(addr)
	storage := s.storage[addr]
	
	if len(storage) == 0 {
		// No storage, set empty root
		acc.Root = [32]byte{}
		return nil
	}
	
	// Create storage trie
	storageTrie, err := trie.NewTrie(acc.Root, s.db)
	if err != nil {
		return err
	}
	
	// Update all storage values
	for key, value := range storage {
		if err := storageTrie.Update(key[:], value[:]); err != nil {
			return err
		}
	}
	
	// Commit storage trie
	root, err := storageTrie.Commit()
	if err != nil {
		return err
	}
	
	acc.Root = root
	return nil
}

// Copy creates a deep copy of the state
func (s *StateDB) Copy() *StateDB {
	newState := &StateDB{
		db:       s.db,
		trie:     s.trie.Copy(),
		accounts: make(map[[20]byte]*Account),
		codes:    make(map[[32]byte][]byte),
		storage:  make(map[[20]byte]map[[32]byte][32]byte),
		logs:     make([]*Log, 0),
		dirty:    make(map[[20]byte]bool),
	}
	
	// Copy accounts
	for addr, acc := range s.accounts {
		newState.accounts[addr] = &Account{
			Nonce:    acc.Nonce,
			Balance:  new(big.Int).Set(acc.Balance),
			CodeHash: acc.CodeHash,
			Root:     acc.Root,
		}
	}
	
	// Copy codes
	for hash, code := range s.codes {
		newState.codes[hash] = make([]byte, len(code))
		copy(newState.codes[hash], code)
	}
	
	// Copy storage
	for addr, storage := range s.storage {
		newState.storage[addr] = make(map[[32]byte][32]byte)
		for key, value := range storage {
			newState.storage[addr][key] = value
		}
	}
	
	// Copy logs
	for _, log := range s.logs {
		newState.logs = append(newState.logs, &Log{
			Address: log.Address,
			Topics:  log.Topics,
			Data:    make([]byte, len(log.Data)),
		})
		copy(newState.logs[len(newState.logs)-1].Data, log.Data)
	}
	
	return newState
}
