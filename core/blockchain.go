package core

import (
	"blockchain-node/cache"
	"blockchain-node/crypto"
	"blockchain-node/database"
	"blockchain-node/interfaces"
	"blockchain-node/logger"
	"blockchain-node/metrics"
	"blockchain-node/state"
	"blockchain-node/validation"
	"errors"
	"fmt"
	"math/big"
	"sync"
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
	currentBlock *Block
	blocks      map[[32]byte]*Block
	blockByNumber map[uint64]*Block
	mempool     *Mempool
	vm          interfaces.VirtualMachine
	consensus   interfaces.Engine
	validator   *validation.Validator
	cache       *cache.Cache
	mu          sync.RWMutex
	shutdownCh  chan struct{}
}

func NewBlockchain(config *Config) (*Blockchain, error) {
	logger.Infof("Initializing custom blockchain with ChainID: %d", config.ChainID)
	
	// Initialize database
	db, err := database.NewLevelDB(config.DataDir + "/chaindata")
	if err != nil {
		logger.Errorf("Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Initialize state database with empty root
	stateDB, err := state.NewStateDB([32]byte{}, db)
	if err != nil {
		logger.Errorf("Failed to create state database: %v", err)
		return nil, fmt.Errorf("failed to create state database: %v", err)
	}

	bc := &Blockchain{
		config:        config,
		db:            db,
		stateDB:       stateDB,
		blocks:        make(map[[32]byte]*Block),
		blockByNumber: make(map[uint64]*Block),
		mempool:       NewMempool(),
		validator:     validation.NewValidator(),
		cache:         cache.NewCache(),
		shutdownCh:    make(chan struct{}),
	}

	// VM will be set later to avoid circular dependency
	bc.vm = nil

	// Load or create genesis block
	if err := bc.initGenesis(); err != nil {
		logger.Errorf("Failed to initialize genesis: %v", err)
		return nil, fmt.Errorf("failed to initialize genesis: %v", err)
	}

	logger.Info("Custom blockchain initialized successfully")
	return bc, nil
}

// SetVirtualMachine sets the virtual machine for the blockchain
func (bc *Blockchain) SetVirtualMachine(vm interfaces.VirtualMachine) {
	bc.vm = vm
}

// SetConsensus sets the consensus engine for the blockchain
func (bc *Blockchain) SetConsensus(consensus interfaces.Engine) {
	bc.consensus = consensus
}

func (bc *Blockchain) initGenesis() error {
	logger.Info("Initializing genesis block")
	
	// Check if genesis block already exists
	if block := bc.GetBlockByNumber(0); block != nil {
		bc.currentBlock = block
		logger.Infof("Genesis block already exists: %x", block.Header.Hash)
		return nil
	}

	// Create genesis block
	genesis := &Block{
		Header: &BlockHeader{
			Number:       0,
			ParentHash:   [32]byte{},
			Timestamp:    1640995200, // Jan 1, 2022
			StateRoot:    [32]byte{},
			TxHash:       [32]byte{},
			ReceiptHash:  [32]byte{},
			GasLimit:     bc.config.BlockGasLimit,
			GasUsed:      0,
			Difficulty:   big.NewInt(1000),
		},
		Transactions: []*Transaction{},
		Receipts:     []*TransactionReceipt{},
	}

	// Set up genesis state (allocate some initial balances)
	genesisAllocation := map[[20]byte]*big.Int{
		[20]byte{0x74, 0x2d, 0x35, 0xcc, 0x66, 0x35, 0xc0, 0x53, 0x29, 0x25, 0xa3, 0xb8, 0xd5, 0xc6, 0xc1, 0xc8, 0xb1, 0xc5, 0xc6, 0xc}: big.NewInt(1e18), // 1 ETH
	}

	for addr, balance := range genesisAllocation {
		bc.stateDB.SetBalance(addr, balance)
		logger.Debugf("Genesis allocation: %x -> %s", addr, balance.String())
	}

	// Commit state and get state root
	stateRoot, err := bc.stateDB.Commit()
	if err != nil {
		logger.Errorf("Failed to commit genesis state: %v", err)
		return fmt.Errorf("failed to commit genesis state: %v", err)
	}

	genesis.Header.StateRoot = stateRoot
	genesis.Header.Hash = genesis.CalculateHash()

	// Save genesis block
	bc.blocks[genesis.Header.Hash] = genesis
	bc.blockByNumber[0] = genesis
	bc.currentBlock = genesis

	// Update metrics
	metrics.GetMetrics().IncrementBlockCount()
	
	logger.BlockEvent(0, fmt.Sprintf("%x", genesis.Header.Hash), 0, "genesis")
	
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

func (bc *Blockchain) GetBlockByHash(hash [32]byte) *Block {
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

	// Validate block using custom validator
	if err := bc.validator.ValidateBlock(block); err != nil {
		logger.Errorf("Block validation failed: %v", err)
		metrics.GetMetrics().IncrementErrorCount()
		return err
	}

	// Validate proof of work if consensus engine is available
	if bc.consensus != nil && !bc.consensus.ValidateProofOfWork(block) {
		logger.Errorf("Invalid proof of work for block %d", block.Header.Number)
		metrics.GetMetrics().IncrementErrorCount()
		return errors.New("invalid proof of work")
	}

	// Execute transactions using custom VM
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
	logger.LogBlockEvent(block.Header.Number, fmt.Sprintf("%x", block.Header.Hash), len(block.Transactions), "miner")

	// Save to database
	if err := bc.saveBlock(block); err != nil {
		logger.Errorf("Failed to save block: %v", err)
		return err
	}

	logger.Infof("Block %d added successfully", block.Header.Number)
	return nil
}

func (bc *Blockchain) executeBlock(block *Block) error {
	logger.Debugf("Executing block %d with %d transactions", block.Header.Number, len(block.Transactions))
	
	// Create new state database for this block
	stateDB, err := state.NewStateDB(bc.currentBlock.Header.StateRoot, bc.db)
	if err != nil {
		return fmt.Errorf("failed to create state database: %v", err)
	}

	var receipts []*TransactionReceipt
	var logs []*Log
	gasUsed := uint64(0)

	// Execute each transaction using custom VM if available
	for i, tx := range block.Transactions {
		logger.Debugf("Executing transaction %d: %x", i, tx.Hash)
		
		// Create execution context
		ctx := &interfaces.ExecutionContext{
			Transaction: tx,
			BlockHeader: block.Header,
			From:        tx.From,
			To:          tx.To,
			Value:       tx.Value,
			Data:        tx.Data,
		}

		var result *interfaces.ExecutionResult
		if bc.vm != nil {
			// Execute transaction with VM
			result, err = bc.vm.ExecuteTransaction(ctx)
			if err != nil {
				logger.Errorf("Failed to execute transaction %d: %v", i, err)
				return fmt.Errorf("failed to execute transaction %d: %v", i, err)
			}
		} else {
			// Simple execution without VM (for basic transactions)
			result = &interfaces.ExecutionResult{
				GasUsed: 21000, // Basic gas cost
				Status:  1,     // Success
				Logs:    []interfaces.ExecutionLog{},
			}
		}

		// Create receipt
		receipt := &TransactionReceipt{
			TxHash:          tx.Hash,
			TxIndex:         uint64(i),
			BlockHash:       block.Header.Hash,
			BlockNumber:     block.Header.Number,
			From:            tx.From,
			To:              tx.To,
			GasUsed:         result.GasUsed,
			CumulativeGasUsed: gasUsed + result.GasUsed,
			Status:          1, // Success
			Logs:            make([]*Log, len(result.Logs)),
		}

		if result.ContractAddress != nil {
			receipt.ContractAddress = result.ContractAddress
		}

		// Convert execution logs to receipt logs
		for j, execLog := range result.Logs {
			receipt.Logs[j] = &Log{
				Address:     execLog.Address,
				Topics:      execLog.Topics,
				Data:        execLog.Data,
				BlockNumber: block.Header.Number,
				TxHash:      tx.Hash,
				TxIndex:     uint64(i),
				BlockHash:   block.Header.Hash,
				Index:       uint64(j),
			}
		}

		receipts = append(receipts, receipt)
		logs = append(logs, receipt.Logs...)
		gasUsed += result.GasUsed
		
		// Update transaction metrics
		metrics.GetMetrics().IncrementTransactionCount()
		
		// Log transaction event
		logger.LogTransactionEvent(
			fmt.Sprintf("%x", tx.Hash),
			fmt.Sprintf("%x", tx.From),
			func() string {
				if tx.To != nil {
					return fmt.Sprintf("%x", *tx.To)
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

	// Commit state changes
	stateRoot, err := stateDB.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit state: %v", err)
	}

	block.Header.StateRoot = stateRoot
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

func (bc *Blockchain) GetBalance(address [20]byte) *big.Int {
	return bc.stateDB.GetBalance(address)
}

func (bc *Blockchain) GetNonce(address [20]byte) uint64 {
	return bc.stateDB.GetNonce(address)
}

func (bc *Blockchain) GetCode(address [20]byte) []byte {
	return bc.stateDB.GetCode(address)
}

func (bc *Blockchain) GetStorageAt(address [20]byte, key [32]byte) [32]byte {
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

func (bc *Blockchain) GetDatabase() database.Database {
	return bc.db
}
