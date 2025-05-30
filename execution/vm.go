
package execution

import (
	"blockchain-node/core"
	"blockchain-node/crypto"
	"blockchain-node/state"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

// Virtual Machine untuk eksekusi transaksi kustom
type VirtualMachine struct {
	stateDB *state.StateDB
	gasUsed uint64
	gasLimit uint64
}

// Gas constants untuk operasi berbeda
const (
	GasTransfer     = 21000  // Gas untuk transfer ETH
	GasCreate       = 32000  // Gas untuk membuat contract
	GasCall         = 2300   // Gas untuk call
	GasStorage      = 20000  // Gas untuk storage operation
	GasComputation  = 3      // Gas per computation step
)

// Instruction opcodes untuk VM kustom
const (
	OpNOP      = 0x00 // No operation
	OpPUSH     = 0x01 // Push value to stack
	OpPOP      = 0x02 // Pop value from stack
	OpADD      = 0x03 // Addition
	OpSUB      = 0x04 // Subtraction
	OpMUL      = 0x05 // Multiplication
	OpDIV      = 0x06 // Division
	OpMOD      = 0x07 // Modulo
	OpSTORE    = 0x08 // Store to storage
	OpLOAD     = 0x09 // Load from storage
	OpBALANCE  = 0x0A // Get balance
	OpTRANSFER = 0x0B // Transfer value
	OpRETURN   = 0x0C // Return from execution
	OpREVERT   = 0x0D // Revert transaction
)

// ExecutionContext berisi konteks eksekusi transaksi
type ExecutionContext struct {
	Transaction *core.Transaction
	BlockHeader *core.BlockHeader
	From        [20]byte
	To          *[20]byte
	Value       *big.Int
	Data        []byte
	GasUsed     uint64
}

// ExecutionResult berisi hasil eksekusi transaksi
type ExecutionResult struct {
	Success         bool
	GasUsed         uint64
	ReturnData      []byte
	ContractAddress *[20]byte
	Logs            []*core.Log
	Error           error
}

// NewVirtualMachine membuat VM baru
func NewVirtualMachine(stateDB *state.StateDB) *VirtualMachine {
	return &VirtualMachine{
		stateDB: stateDB,
	}
}

// ExecuteTransaction mengeksekusi transaksi dalam VM
func (vm *VirtualMachine) ExecuteTransaction(ctx *ExecutionContext) (*ExecutionResult, error) {
	// Reset gas counter
	vm.gasUsed = 0
	vm.gasLimit = ctx.Transaction.GasLimit
	
	result := &ExecutionResult{
		Success: false,
		Logs:    make([]*core.Log, 0),
	}
	
	// Charge base gas
	if !vm.consumeGas(GasTransfer) {
		return result, errors.New("insufficient gas for transaction")
	}
	
	// Validate transaction
	if err := vm.validateTransaction(ctx); err != nil {
		result.Error = err
		return result, err
	}
	
	// Check if this is a contract creation or call
	if ctx.To == nil {
		// Contract creation
		return vm.executeContractCreation(ctx)
	} else {
		// Regular transaction or contract call
		return vm.executeCall(ctx)
	}
}

// validateTransaction melakukan validasi transaksi
func (vm *VirtualMachine) validateTransaction(ctx *ExecutionContext) error {
	// Check balance
	fromBalance := vm.stateDB.GetBalance(ctx.From)
	totalCost := new(big.Int).Add(ctx.Value, new(big.Int).Mul(
		ctx.Transaction.GasPrice, 
		big.NewInt(int64(ctx.Transaction.GasLimit)),
	))
	
	if fromBalance.Cmp(totalCost) < 0 {
		return errors.New("insufficient balance")
	}
	
	// Check nonce
	currentNonce := vm.stateDB.GetNonce(ctx.From)
	if ctx.Transaction.Nonce != currentNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", currentNonce, ctx.Transaction.Nonce)
	}
	
	return nil
}

// executeCall mengeksekusi panggilan ke address yang ada
func (vm *VirtualMachine) executeCall(ctx *ExecutionContext) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Logs: make([]*core.Log, 0),
	}
	
	toAddress := *ctx.To
	
	// Transfer value jika ada
	if ctx.Value.Cmp(big.NewInt(0)) > 0 {
		if !vm.consumeGas(GasTransfer) {
			result.Error = errors.New("insufficient gas for transfer")
			return result, result.Error
		}
		
		// Perform transfer
		fromBalance := vm.stateDB.GetBalance(ctx.From)
		toBalance := vm.stateDB.GetBalance(toAddress)
		
		newFromBalance := new(big.Int).Sub(fromBalance, ctx.Value)
		newToBalance := new(big.Int).Add(toBalance, ctx.Value)
		
		vm.stateDB.SetBalance(ctx.From, newFromBalance)
		vm.stateDB.SetBalance(toAddress, newToBalance)
	}
	
	// Execute contract code jika ada data
	if len(ctx.Data) > 0 {
		code := vm.stateDB.GetCode(toAddress)
		if len(code) > 0 {
			// Execute contract code
			returnData, err := vm.executeCode(code, ctx.Data, ctx)
			if err != nil {
				result.Error = err
				return result, err
			}
			result.ReturnData = returnData
		} else {
			// Simple data storage untuk non-contract address
			if err := vm.executeSimpleData(ctx); err != nil {
				result.Error = err
				return result, err
			}
		}
	}
	
	// Update nonce
	vm.stateDB.SetNonce(ctx.From, ctx.Transaction.Nonce+1)
	
	result.Success = true
	result.GasUsed = vm.gasUsed
	return result, nil
}

// executeContractCreation mengeksekusi pembuatan contract baru
func (vm *VirtualMachine) executeContractCreation(ctx *ExecutionContext) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Logs: make([]*core.Log, 0),
	}
	
	// Charge gas untuk creation
	if !vm.consumeGas(GasCreate) {
		result.Error = errors.New("insufficient gas for contract creation")
		return result, result.Error
	}
	
	// Generate contract address
	contractAddr := vm.generateContractAddress(ctx.From, ctx.Transaction.Nonce)
	
	// Check if address is available
	if vm.stateDB.GetBalance(contractAddr).Cmp(big.NewInt(0)) > 0 || 
	   vm.stateDB.GetNonce(contractAddr) > 0 {
		result.Error = errors.New("contract address collision")
		return result, result.Error
	}
	
	// Set initial balance jika ada value
	if ctx.Value.Cmp(big.NewInt(0)) > 0 {
		fromBalance := vm.stateDB.GetBalance(ctx.From)
		newFromBalance := new(big.Int).Sub(fromBalance, ctx.Value)
		
		vm.stateDB.SetBalance(ctx.From, newFromBalance)
		vm.stateDB.SetBalance(contractAddr, ctx.Value)
	}
	
	// Deploy contract code
	if len(ctx.Data) > 0 {
		vm.stateDB.SetCode(contractAddr, ctx.Data)
	}
	
	// Update nonce
	vm.stateDB.SetNonce(ctx.From, ctx.Transaction.Nonce+1)
	vm.stateDB.SetNonce(contractAddr, 1) // Contract nonce starts at 1
	
	result.Success = true
	result.GasUsed = vm.gasUsed
	result.ContractAddress = &contractAddr
	return result, nil
}

// executeCode mengeksekusi bytecode menggunakan VM kustom
func (vm *VirtualMachine) executeCode(code []byte, input []byte, ctx *ExecutionContext) ([]byte, error) {
	// Simple stack-based VM
	stack := make([]*big.Int, 0, 1024)
	storage := make(map[[32]byte]*big.Int)
	pc := 0 // Program counter
	
	for pc < len(code) {
		if !vm.consumeGas(GasComputation) {
			return nil, errors.New("out of gas")
		}
		
		opcode := code[pc]
		pc++
		
		switch opcode {
		case OpNOP:
			// Do nothing
			
		case OpPUSH:
			if pc+32 > len(code) {
				return nil, errors.New("invalid PUSH instruction")
			}
			value := new(big.Int).SetBytes(code[pc:pc+32])
			stack = append(stack, value)
			pc += 32
			
		case OpPOP:
			if len(stack) == 0 {
				return nil, errors.New("stack underflow")
			}
			stack = stack[:len(stack)-1]
			
		case OpADD:
			if len(stack) < 2 {
				return nil, errors.New("insufficient values for ADD")
			}
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			result := new(big.Int).Add(a, b)
			stack = stack[:len(stack)-2]
			stack = append(stack, result)
			
		case OpSUB:
			if len(stack) < 2 {
				return nil, errors.New("insufficient values for SUB")
			}
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			result := new(big.Int).Sub(b, a) // b - a
			stack = stack[:len(stack)-2]
			stack = append(stack, result)
			
		case OpMUL:
			if len(stack) < 2 {
				return nil, errors.New("insufficient values for MUL")
			}
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			result := new(big.Int).Mul(a, b)
			stack = stack[:len(stack)-2]
			stack = append(stack, result)
			
		case OpDIV:
			if len(stack) < 2 {
				return nil, errors.New("insufficient values for DIV")
			}
			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			if a.Cmp(big.NewInt(0)) == 0 {
				return nil, errors.New("division by zero")
			}
			result := new(big.Int).Div(b, a)
			stack = stack[:len(stack)-2]
			stack = append(stack, result)
			
		case OpSTORE:
			if !vm.consumeGas(GasStorage) {
				return nil, errors.New("out of gas for storage")
			}
			if len(stack) < 2 {
				return nil, errors.New("insufficient values for STORE")
			}
			key := stack[len(stack)-1]
			value := stack[len(stack)-2]
			
			var keyBytes [32]byte
			copy(keyBytes[:], key.Bytes())
			storage[keyBytes] = value
			
			stack = stack[:len(stack)-2]
			
		case OpLOAD:
			if len(stack) < 1 {
				return nil, errors.New("insufficient values for LOAD")
			}
			key := stack[len(stack)-1]
			
			var keyBytes [32]byte
			copy(keyBytes[:], key.Bytes())
			
			value, exists := storage[keyBytes]
			if !exists {
				value = big.NewInt(0)
			}
			
			stack[len(stack)-1] = value
			
		case OpRETURN:
			if len(stack) > 0 {
				returnValue := stack[len(stack)-1]
				return returnValue.Bytes(), nil
			}
			return []byte{}, nil
			
		case OpREVERT:
			return nil, errors.New("execution reverted")
			
		default:
			return nil, fmt.Errorf("unknown opcode: 0x%02x", opcode)
		}
	}
	
	// Eksekusi selesai tanpa RETURN
	return []byte{}, nil
}

// executeSimpleData mengeksekusi data sederhana (bukan contract code)
func (vm *VirtualMachine) executeSimpleData(ctx *ExecutionContext) error {
	// Simple interpretation: data as key-value pairs
	if len(ctx.Data)%64 == 0 { // 32 bytes key + 32 bytes value
		for i := 0; i < len(ctx.Data); i += 64 {
			key := ctx.Data[i:i+32]
			value := ctx.Data[i+32:i+64]
			
			var keyHash [32]byte
			var valueHash [32]byte
			copy(keyHash[:], key)
			copy(valueHash[:], value)
			
			if !vm.consumeGas(GasStorage) {
				return errors.New("out of gas for storage")
			}
			
			vm.stateDB.SetState(*ctx.To, keyHash, valueHash)
		}
	}
	
	return nil
}

// generateContractAddress menggenerate address untuk contract baru
func (vm *VirtualMachine) generateContractAddress(creator [20]byte, nonce uint64) [20]byte {
	// Use creator address + nonce to generate contract address
	data := append(creator[:], big.NewInt(int64(nonce)).Bytes()...)
	hash := crypto.Keccak256Hash(data)
	
	var addr [20]byte
	copy(addr[:], hash[12:]) // Take last 20 bytes
	return addr
}

// consumeGas mengkonsumsi gas untuk operasi
func (vm *VirtualMachine) consumeGas(amount uint64) bool {
	if vm.gasUsed+amount > vm.gasLimit {
		return false
	}
	vm.gasUsed += amount
	return true
}

// GetGasUsed mengembalikan total gas yang digunakan
func (vm *VirtualMachine) GetGasUsed() uint64 {
	return vm.gasUsed
}
