
package validation

import (
	"blockchain-node/core"
	"blockchain-node/logger"
	"errors"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
)

type Validator struct {
	maxTransactionSize  uint64
	maxBlockSize        uint64
	maxGasLimit         uint64
	minGasPrice         *big.Int
	addressRegex        *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		maxTransactionSize: 128 * 1024,      // 128 KB
		maxBlockSize:       1024 * 1024,     // 1 MB
		maxGasLimit:        10000000,        // 10M gas
		minGasPrice:        big.NewInt(1000), // 1000 wei minimum
		addressRegex:       regexp.MustCompile("^0x[a-fA-F0-9]{40}$"),
	}
}

func (v *Validator) ValidateTransaction(tx *core.Transaction) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	
	// Validate gas price
	if tx.GasPrice == nil || tx.GasPrice.Cmp(v.minGasPrice) < 0 {
		logger.Warningf("Transaction gas price too low: %v", tx.GasPrice)
		return errors.New("gas price too low")
	}
	
	// Validate gas limit
	if tx.GasLimit == 0 || tx.GasLimit > v.maxGasLimit {
		logger.Warningf("Invalid gas limit: %d", tx.GasLimit)
		return errors.New("invalid gas limit")
	}
	
	// Validate value
	if tx.Value == nil || tx.Value.Sign() < 0 {
		logger.Warningf("Invalid transaction value: %v", tx.Value)
		return errors.New("invalid transaction value")
	}
	
	// Validate to address format if present
	if tx.To != nil && !v.IsValidAddress(tx.To.Hex()) {
		logger.Warningf("Invalid to address: %s", tx.To.Hex())
		return errors.New("invalid to address")
	}
	
	// Validate from address
	if tx.From == (common.Address{}) {
		logger.Warning("Transaction missing from address")
		return errors.New("missing from address")
	}
	
	if !v.IsValidAddress(tx.From.Hex()) {
		logger.Warningf("Invalid from address: %s", tx.From.Hex())
		return errors.New("invalid from address")
	}
	
	// Validate signature components
	if tx.V == nil || tx.R == nil || tx.S == nil {
		logger.Warning("Transaction missing signature components")
		return errors.New("missing signature components")
	}
	
	// Validate transaction size
	txData, err := tx.ToJSON()
	if err != nil {
		logger.Errorf("Failed to serialize transaction: %v", err)
		return errors.New("failed to serialize transaction")
	}
	
	if uint64(len(txData)) > v.maxTransactionSize {
		logger.Warningf("Transaction size too large: %d bytes", len(txData))
		return errors.New("transaction size too large")
	}
	
	// Verify signature
	if !tx.VerifySignature() {
		logger.Warning("Invalid transaction signature")
		return errors.New("invalid transaction signature")
	}
	
	logger.Debugf("Transaction validation passed: %s", tx.Hash.Hex())
	return nil
}

func (v *Validator) ValidateBlock(block *core.Block) error {
	if block == nil {
		return errors.New("block is nil")
	}
	
	if block.Header == nil {
		return errors.New("block header is nil")
	}
	
	// Validate block gas limit
	if block.Header.GasLimit > v.maxGasLimit {
		logger.Warningf("Block gas limit too high: %d", block.Header.GasLimit)
		return errors.New("block gas limit too high")
	}
	
	// Validate gas used doesn't exceed limit
	if block.Header.GasUsed > block.Header.GasLimit {
		logger.Warningf("Block gas used exceeds limit: %d > %d", block.Header.GasUsed, block.Header.GasLimit)
		return errors.New("block gas used exceeds limit")
	}
	
	// Validate block timestamp (should not be too far in future)
	// Allow up to 15 minutes in future
	if block.Header.Timestamp > (getCurrentTimestamp() + 900) {
		logger.Warningf("Block timestamp too far in future: %d", block.Header.Timestamp)
		return errors.New("block timestamp too far in future")
	}
	
	// Validate block size
	blockData, err := block.ToJSON()
	if err != nil {
		logger.Errorf("Failed to serialize block: %v", err)
		return errors.New("failed to serialize block")
	}
	
	if uint64(len(blockData)) > v.maxBlockSize {
		logger.Warningf("Block size too large: %d bytes", len(blockData))
		return errors.New("block size too large")
	}
	
	// Validate all transactions in block
	totalGasUsed := uint64(0)
	for i, tx := range block.Transactions {
		if err := v.ValidateTransaction(tx); err != nil {
			logger.Errorf("Invalid transaction %d in block: %v", i, err)
			return err
		}
		totalGasUsed += tx.GasLimit
	}
	
	// Check if calculated gas matches header
	if totalGasUsed != block.Header.GasUsed {
		logger.Warningf("Block gas used mismatch: calculated %d, header %d", totalGasUsed, block.Header.GasUsed)
		return errors.New("block gas used mismatch")
	}
	
	logger.Debugf("Block validation passed: %s", block.Header.Hash.Hex())
	return nil
}

func (v *Validator) IsValidAddress(address string) bool {
	return v.addressRegex.MatchString(address)
}

func (v *Validator) ValidateGasPrice(gasPrice *big.Int) bool {
	return gasPrice != nil && gasPrice.Cmp(v.minGasPrice) >= 0
}

func (v *Validator) ValidateGasLimit(gasLimit uint64) bool {
	return gasLimit > 0 && gasLimit <= v.maxGasLimit
}

func getCurrentTimestamp() int64 {
	return int64(1640995200) // Placeholder for current timestamp
}
