
package evm

import (
	"blockchain-node/core"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type EVM struct {
	blockchain Blockchain
	vmConfig   vm.Config
}

type Blockchain interface {
	GetConfig() interface{}
	GetCurrentBlock() *core.Block
	GetBlockByHash(hash [32]byte) *core.Block
	GetBlockByNumber(number uint64) *core.Block
}

func NewEVM(blockchain Blockchain) *EVM {
	return &EVM{
		blockchain: blockchain,
		vmConfig: vm.Config{
			Debug: false,
		},
	}
}

func (e *EVM) ExecuteTransaction(stateDB *state.StateDB, tx *core.Transaction, header *core.BlockHeader, gasUsed uint64) (*core.TransactionReceipt, error) {
	// Create EVM context
	context := vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     e.GetHashFn(header),
		Coinbase:    common.Address{}, // Miner address
		BlockNumber: new(big.Int).SetUint64(header.Number),
		Time:        new(big.Int).SetInt64(header.Timestamp),
		Difficulty:  header.Difficulty,
		GasLimit:    header.GasLimit,
	}

	// Create transaction context
	txContext := vm.TxContext{
		Origin:   tx.From,
		GasPrice: tx.GasPrice,
	}

	// Create EVM instance
	evm := vm.NewEVM(context, txContext, stateDB, params.MainnetChainConfig, e.vmConfig)

	// Convert our transaction to Ethereum transaction
	ethTx := tx.ToEthTransaction()

	// Execute transaction
	var (
		result *vm.ExecutionResult
		err    error
	)

	if tx.IsContractCreation() {
		// Contract creation
		result, _, err = evm.Create(vm.AccountRef(tx.From), tx.Data, tx.GasLimit, tx.Value)
	} else {
		// Contract call or simple transfer
		result, err = evm.Call(vm.AccountRef(tx.From), *tx.To, tx.Data, tx.GasLimit, tx.Value)
	}

	// Create receipt
	receipt := &core.TransactionReceipt{
		TxHash:      tx.Hash,
		TxIndex:     0, // Will be set by caller
		BlockHash:   header.Hash,
		BlockNumber: header.Number,
		From:        tx.From,
		To:          tx.To,
		GasUsed:     tx.GasLimit - result.LeftOverGas,
		Logs:        convertLogs(result.Logs, header, tx),
		Status:      1, // Success
	}

	if err != nil {
		receipt.Status = 0 // Failed
	}

	// Set contract address for contract creation
	if tx.IsContractCreation() && err == nil {
		contractAddr := crypto.CreateAddress(tx.From, tx.Nonce)
		receipt.ContractAddress = &contractAddr
	}

	return receipt, nil
}

func (e *EVM) GetHashFn(header *core.BlockHeader) vm.GetHashFunc {
	return func(n uint64) common.Hash {
		if block := e.blockchain.GetBlockByNumber(n); block != nil {
			return common.BytesToHash(block.Header.Hash[:])
		}
		return common.Hash{}
	}
}

func convertLogs(vmLogs []*ethTypes.Log, header *core.BlockHeader, tx *core.Transaction) []*core.Log {
	var logs []*core.Log
	for i, vmLog := range vmLogs {
		log := &core.Log{
			Address:     vmLog.Address,
			Topics:      vmLog.Topics,
			Data:        vmLog.Data,
			BlockNumber: header.Number,
			TxHash:      tx.Hash,
			TxIndex:     0, // Will be set by caller
			BlockHash:   header.Hash,
			Index:       uint64(i),
			Removed:     false,
		}
		logs = append(logs, log)
	}
	return logs
}

// Helper functions for EVM
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
