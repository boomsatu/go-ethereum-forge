
package execution

import (
	"blockchain-node/interfaces"
	"blockchain-node/state"
	"fmt"
	"math/big"
)

type VirtualMachine struct {
	stateDB *state.StateDB
}

func NewVirtualMachine(stateDB *state.StateDB) *VirtualMachine {
	return &VirtualMachine{
		stateDB: stateDB,
	}
}

func (vm *VirtualMachine) ExecuteTransaction(ctx *interfaces.ExecutionContext) (*interfaces.ExecutionResult, error) {
	// Simple transaction execution
	// In a real implementation, this would handle smart contracts, etc.
	
	// Basic gas cost for simple transfer
	gasUsed := uint64(21000)
	
	// For contract creation, add extra gas
	if ctx.To == nil && len(ctx.Data) > 0 {
		gasUsed += uint64(len(ctx.Data)) * 68
	}
	
	// Update balances for simple transfers
	if ctx.Value.Cmp(big.NewInt(0)) > 0 {
		// Check if sender has enough balance
		senderBalance := vm.stateDB.GetBalance(ctx.From)
		if senderBalance.Cmp(ctx.Value) < 0 {
			return &interfaces.ExecutionResult{
				GasUsed: gasUsed,
				Status:  0, // Failed
				Error:   ErrInsufficientBalance,
			}, nil
		}
		
		// Transfer funds
		vm.stateDB.SubBalance(ctx.From, ctx.Value)
		if ctx.To != nil {
			vm.stateDB.AddBalance(*ctx.To, ctx.Value)
		}
	}
	
	return &interfaces.ExecutionResult{
		GasUsed: gasUsed,
		Status:  1, // Success
		Logs:    []interfaces.ExecutionLog{},
	}, nil
}

// Execution errors
var (
	ErrInsufficientBalance = fmt.Errorf("insufficient balance")
	ErrInvalidTransaction  = fmt.Errorf("invalid transaction")
	ErrContractFailed      = fmt.Errorf("contract execution failed")
)
