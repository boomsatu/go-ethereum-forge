package core

import (
	"blockchain-node/cache"
	"blockchain-node/crypto"
	"blockchain-node/database"
	"blockchain-node/evm"
	"blockchain-node/logger"
	"blockchain-node/metrics"
	"blockchain-node/validation"
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
	validator   *validation.Validator
	cache       *cache.Cache
	mu          sync.RWMutex
	shutdownCh  chan struct{}
}

func NewBlockchain(config *Config) (*Blockchain, error) {
	logger.Infof("Initializing blockchain with ChainID: %d", config.ChainID)
	
	// Initialize database
	db, err := database.NewLevelDB(config.DataDir + "/chaindata")
	if err != nil {
		logger.Errorf("Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Initialize trie database
	trieDB := trie.NewDatabase(db.GetEthDB())

	// Initialize state database
	stateDB, err := state.New(common.Hash{}, state.NewDatabase(trieDB), nil)
	if err != nil {
		logger.Errorf("Failed to create state database: %v", err)
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
		validator:     validation.NewValidator(),
		cache:         cache.NewCache(),
		shutdownCh:    make(chan struct{}),
	}

	// Initialize EVM
	bc.evm = evm.NewEVM(bc)

	// Load or create genesis block
	if err := bc.initGenesis(); err != nil {
		logger.Errorf("Failed to initialize genesis: %v", err)
		return nil, fmt.Errorf("failed to initialize genesis: %v", err)
	}

	logger.Info("Blockchain initialized successfully")
	return bc, nil
}

func (bc *Blockchain) initGenesis() error {
	logger.Info("Initializing genesis block")
	
	// Check if genesis block already exists
	if block := bc.GetBlockByNumber(0); block != nil {
		bc.currentBlock = block
		logger.Infof("Genesis block already exists: %s", block.Header.Hash.Hex())
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
		logger.Debugf("Genesis allocation: %s -> %s", addr.Hex(), balance.String())
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
		logger.Errorf("Failed to commit genesis state: %v", err)
		return fmt.Errorf("failed to commit genesis state: %v", err)
	}

	if err := bc.trieDB.Commit(root, false, nil); err != nil {
		logger.Errorf("Failed to commit genesis trie: %v", err)
		return fmt.Errorf("failed to commit genesis trie: %v", err)
	}

	// Update metrics
	metrics.GetMetrics().IncrementBlockCount()
	
	logger.BlockEvent(0, genesis.Header.Hash.Hex(), 0, "genesis")
	
	if err := bc.saveBlock(genesis); err != nil {
		logger.Errorf("Failed to save genesis block: %v", err)
		return err
	}
	
	logger.Info("Genesis block created successfully")
	return nil
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

	logger.Debugf("Adding block %d to blockchain", block.Header.Number)

	// Validate block
	if err := bc.validator.ValidateBlock(block); err != nil {
		logger.Errorf("Block validation failed: %v", err)
		metrics.GetMetrics().IncrementErrorCount()
		return err
	}

	// Execute transactions
	if err := bc.executeBlock(block); err != nil {
		logger.Errorf("Block execution failed: %v", err)
		metrics.GetMetrics().IncrementErrorCount()
		return err
	}

	// Add to blockchain
	bc.blocks[block.Header.Hash] = block
	bc.blockByNumber[block.Header.Number] = block
	bc.currentBlock = block

	// Update metrics
	metrics.GetMetrics().IncrementBlockCount()
	metrics.GetMetrics().SetTransactionPoolSize(uint32(bc.mempool.GetPendingCount()))

	// Log block event
	logger.LogBlockEvent(block.Header.Number, block.Header.Hash.Hex(), len(block.Transactions), "miner")

	// Save to database
	if err := bc.saveBlock(block); err != nil {
		logger.Errorf("Failed to save block: %v", err)
		return err
	}

	logger.Infof("Block %d added successfully", block.Header.Number)
	return nil
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
	logger.Debugf("Executing block %d with %d transactions", block.Header.Number, len(block.Transactions))
	
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
		logger.Debugf("Executing transaction %d: %s", i, tx.Hash.Hex())
		
		receipt, err := bc.evm.ExecuteTransaction(stateDB, tx, block.Header, gasUsed)
		if err != nil {
			logger.Errorf("Failed to execute transaction %d: %v", i, err)
			return fmt.Errorf("failed to execute transaction %d: %v", i, err)
		}

		receipts = append(receipts, receipt)
		logs = append(logs, receipt.Logs...)
		gasUsed += receipt.GasUsed
		
		// Update transaction metrics
		metrics.GetMetrics().IncrementTransactionCount()
		
		// Log transaction event
		logger.LogTransactionEvent(
			tx.Hash.Hex(),
			tx.From.Hex(),
			func() string {
				if tx.To != nil {
					return tx.To.Hex()
				}
				return "contract_creation"
			}(),
			tx.Value.String(),
			"success",
		)

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
	logger.Debugf("Block %d executed successfully", block.Header.Number)
	return nil
}

func (bc *Blockchain) saveBlock(block *Block) error {
	// Implement block serialization and storage
	blockData, err := block.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize block: %v", err)
	}
	
	blockKey := fmt.Sprintf("block_%d", block.Header.Number)
	if err := bc.db.Put([]byte(blockKey), blockData); err != nil {
		return fmt.Errorf("failed to save block: %v", err)
	}
	
	// Cache the block
	bc.cache.Set(blockKey, block, cache.DefaultTTL)
	
	return nil
}

func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	logger.Debugf("Adding transaction to mempool: %s", tx.Hash.Hex())
	
	// Validate transaction
	if err := bc.validator.ValidateTransaction(tx); err != nil {
		logger.Errorf("Transaction validation failed: %v", err)
		return err
	}
	
	if err := bc.mempool.AddTransaction(tx); err != nil {
		logger.Errorf("Failed to add transaction to mempool: %v", err)
		return err
	}
	
	// Update metrics
	metrics.GetMetrics().SetTransactionPoolSize(uint32(bc.mempool.GetPendingCount()))
	
	logger.Debugf("Transaction added to mempool successfully: %s", tx.Hash.Hex())
	return nil
}

func (bc *Blockchain) GetMempool() *Mempool {
	return bc.mempool
}

func (bc *Blockchain) Close() error {
	logger.Info("Closing blockchain")
	
	close(bc.shutdownCh)
	
	if err := bc.db.Close(); err != nil {
		logger.Errorf("Failed to close database: %v", err)
		return err
	}
	
	logger.Info("Blockchain closed successfully")
	return nil
}

func (bc *Blockchain) GetBalance(address common.Address) *big.Int {
	return bc.stateDB.GetBalance(address)
}

func (bc *Blockchain) GetNonce(address common.Address) uint64 {
	return bc.stateDB.GetNonce(address)
}

func (bc *Blockchain) GetCode(address common.Address) []byte {
	return bc.stateDB.GetCode(address)
}

func (bc *Blockchain) GetStorageAt(address common.Address, key common.Hash) common.Hash {
	return bc.stateDB.GetState(address, key)
}

func (bc *Blockchain) EstimateGas(tx *Transaction) (uint64, error) {
	// Simple gas estimation - in production this would be more sophisticated
	baseGas := uint64(21000)
	if len(tx.Data) > 0 {
		baseGas += uint64(len(tx.Data)) * 68
	}
	return baseGas, nil
}
