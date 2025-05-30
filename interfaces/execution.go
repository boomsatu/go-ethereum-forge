
package interfaces

import (
	"math/big"
)

// ExecutionContext represents the context for transaction execution
type ExecutionContext struct {
	Transaction interface{}
	BlockHeader interface{}
	From        [20]byte
	To          *[20]byte
	Value       *big.Int
	Data        []byte
}

// ExecutionResult represents the result of transaction execution
type ExecutionResult struct {
	GasUsed         uint64
	ContractAddress *[20]byte
	Logs            []ExecutionLog
	Status          uint64
	Error           error
}

// ExecutionLog represents a log entry from contract execution
type ExecutionLog struct {
	Address [20]byte
	Topics  [][32]byte
	Data    []byte
}

// VirtualMachine represents the interface for transaction execution
type VirtualMachine interface {
	ExecuteTransaction(ctx *ExecutionContext) (*ExecutionResult, error)
}
