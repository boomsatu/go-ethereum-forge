
package core

import (
	"blockchain-node/crypto"
	"blockchain-node/database"
	"blockchain-node/evm"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type Config struct {
	DataDir       string
	ChainID       uint64
	BlockGasLimit uint64
}

type Blockchain struct {
	config      *Config
	db          database.Database
	stateDB     *state.StateDB
	trieDB      *trie.Database
	currentBlock *Block
	blocks      map[common.Hash]*Block
	blockByNumber map[uint64]*Block
	mempool     *Mempool
	evm         *evm.EVM
	mu          sync.RWMutex
}

func NewBlockchain(config *Config) (*Blockchain, error) {
	// Initialize database
	db, err := database.NewLevelDB(config.DataDir + "/chaindata")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Initialize trie database
	trieDB := trie.NewDatabase(db.GetEthDB())

	// Initialize state database
	stateDB, err := state.New(common.Hash{}, state.NewDatabase(trieDB), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create state database: %v", err)
	}

	bc := &Blockchain{
		config:        config,
		db:            db,
		stateDB:       stateDB,
		trieDB:        trieDB,
		blocks:        make(map[common.Hash]*Block),
		blockByNumber: make(map[uint64]*Block),
		mempool:       NewMempool(),
	}

	// Initialize EVM
	bc.evm = evm.NewEVM(bc)

	// Load or create genesis block
	if err := bc.initGenesis(); err != nil {
		return nil, fmt.Errorf("failed to initialize genesis: %v", err)
	}

	return bc, nil
}

func (bc *Blockchain) initGenesis() error {
	// Check if genesis block already exists
	if block := bc.GetBlockByNumber(0); block != nil {
		bc.currentBlock = block
		return nil
	}

	// Create genesis block
	genesis := &Block{
		Header: &BlockHeader{
			Number:       0,
			ParentHash:   common.Hash{},
			Timestamp:    1640995200, // Jan 1, 2022
			StateRoot:    bc.stateDB.IntermediateRoot(false),
			TxHash:       ethTypes.EmptyRootHash,
			ReceiptHash:  ethTypes.EmptyRootHash,
			GasLimit:     bc.config.BlockGasLimit,
			GasUsed:      0,
			Difficulty:   big.NewInt(1000),
		},
		Transactions: []*Transaction{},
		Receipts:     []*TransactionReceipt{},
	}

	// Set up genesis state (allocate some initial balances)
	genesisAllocation := map[common.Address]*big.Int{
		common.HexToAddress("0x742d35Cc6635C0532925a3b8D5c6C1C8b1c5C6C"): big.NewInt(1e18), // 1 ETH
	}

	for addr, balance := range genesisAllocation {
		bc.stateDB.SetBalance(addr, balance)
	}

	genesis.Header.StateRoot = bc.stateDB.IntermediateRoot(false)
	genesis.Header.Hash = genesis.CalculateHash()

	// Save genesis block
	bc.blocks[genesis.Header.Hash] = genesis
	bc.blockByNumber[0] = genesis
	bc.currentBlock = genesis

	// Commit state
	root, err := bc.stateDB.Commit(false)
	if err != nil {
		return fmt.Errorf("failed to commit genesis state: %v", err)
	}

	if err := bc.trieDB.Commit(root, false, nil); err != nil {
		return fmt.Errorf("failed to commit genesis trie: %v", err)
	}

	return bc.saveBlock(genesis)
}

func (bc *Blockchain) GetConfig() *Config {
	return bc.config
}

func (bc *Blockchain) GetStateDB() *state.StateDB {
	return bc.stateDB
}

func (bc *Blockchain) GetCurrentBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentBlock
}

func (bc *Blockchain) GetBlockByHash(hash common.Hash) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks[hash]
}

func (bc *Blockchain) GetBlockByNumber(number uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blockByNumber[number]
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate block
	if err := bc.validateBlock(block); err != nil {
		return err
	}

	// Execute transactions
	if err := bc.executeBlock(block); err != nil {
		return err
	}

	// Add to blockchain
	bc.blocks[block.Header.Hash] = block
	bc.blockByNumber[block.Header.Number] = block
	bc.currentBlock = block

	// Save to database
	return bc.saveBlock(block)
}

func (bc *Blockchain) validateBlock(block *Block) error {
	if block.Header.Number != bc.currentBlock.Header.Number+1 {
		return errors.New("invalid block number")
	}

	if block.Header.ParentHash != bc.currentBlock.Header.Hash {
		return errors.New("invalid parent hash")
	}

	// Validate proof of work
	if !crypto.ValidateProofOfWork(block.Header.Hash, block.Header.Nonce, block.Header.Difficulty) {
		return errors.New("invalid proof of work")
	}

	return nil
}

func (bc *Blockchain) executeBlock(block *Block) error {
	// Create new state database for this block
	stateDB, err := state.New(bc.currentBlock.Header.StateRoot, state.NewDatabase(bc.trieDB), nil)
	if err != nil {
		return fmt.Errorf("failed to create state database: %v", err)
	}

	var receipts []*TransactionReceipt
	var logs []*Log
	gasUsed := uint64(0)

	// Execute each transaction
	for i, tx := range block.Transactions {
		receipt, err := bc.evm.ExecuteTransaction(stateDB, tx, block.Header, gasUsed)
		if err != nil {
			return fmt.Errorf("failed to execute transaction %d: %v", i, err)
		}

		receipts = append(receipts, receipt)
		logs = append(logs, receipt.Logs...)
		gasUsed += receipt.GasUsed

		if gasUsed > block.Header.GasLimit {
			return errors.New("block gas limit exceeded")
		}
	}

	// Update block with receipts
	block.Receipts = receipts
	block.Header.GasUsed = gasUsed
	block.Header.StateRoot = stateDB.IntermediateRoot(false)

	// Commit state
	root, err := stateDB.Commit(false)
	if err != nil {
		return fmt.Errorf("failed to commit state: %v", err)
	}

	if err := bc.trieDB.Commit(root, false, nil); err != nil {
		return fmt.Errorf("failed to commit trie: %v", err)
	}

	bc.stateDB = stateDB
	return nil
}

func (bc *Blockchain) saveBlock(block *Block) error {
	// Implement block serialization and storage
	return nil
}

func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	return bc.mempool.AddTransaction(tx)
}

func (bc *Blockchain) GetMempool() *Mempool {
	return bc.mempool
}

func (bc *Blockchain) Close() error {
	return bc.db.Close()
}
