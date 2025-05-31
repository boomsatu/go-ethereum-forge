
package validation

import (
	"blockchain-node/logger"
	"errors"
	"math/big"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Validator struct {
	maxTransactionSize  uint64
	maxBlockSize        uint64
	maxGasLimit         uint64
	minGasPrice         *big.Int
	addressRegex        *regexp.Regexp
}

// Transaction interface to avoid circular import
type Transaction interface {
	GetHash() [32]byte
	GetFrom() common.Address
	GetTo() *common.Address
	GetValue() *big.Int
	GetGasPrice() *big.Int
	GetGasLimit() uint64
	GetData() []byte
	GetV() *big.Int
	GetR() *big.Int
	GetS() *big.Int
	VerifySignature() bool
	ToJSON() ([]byte, error)
}

// Block interface to avoid circular import
type Block interface {
	GetHeader() BlockHeader
	GetTransactions() []Transaction
	ToJSON() ([]byte, error)
}

// BlockHeader interface to avoid circular import
type BlockHeader interface {
	GetNumber() uint64
	GetParentHash() [32]byte
	GetTimestamp() int64
	GetGasLimit() uint64
	GetGasUsed() uint64
	GetHash() [32]byte
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

func (v *Validator) ValidateTransaction(tx Transaction) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	
	// Validate gas price
	gasPrice := tx.GetGasPrice()
	if gasPrice == nil || gasPrice.Cmp(v.minGasPrice) < 0 {
		logger.Warningf("Transaction gas price too low: %v", gasPrice)
		return errors.New("gas price too low")
	}
	
	// Validate gas limit
	gasLimit := tx.GetGasLimit()
	if gasLimit == 0 || gasLimit > v.maxGasLimit {
		logger.Warningf("Invalid gas limit: %d", gasLimit)
		return errors.New("invalid gas limit")
	}
	
	// Validate value
	value := tx.GetValue()
	if value == nil || value.Sign() < 0 {
		logger.Warningf("Invalid transaction value: %v", value)
		return errors.New("invalid transaction value")
	}
	
	// Validate to address format if present
	to := tx.GetTo()
	if to != nil && !v.IsValidAddress(to.Hex()) {
		logger.Warningf("Invalid to address: %s", to.Hex())
		return errors.New("invalid to address")
	}
	
	// Validate from address
	from := tx.GetFrom()
	if from == (common.Address{}) {
		logger.Warning("Transaction missing from address")
		return errors.New("missing from address")
	}
	
	if !v.IsValidAddress(from.Hex()) {
		logger.Warningf("Invalid from address: %s", from.Hex())
		return errors.New("invalid from address")
	}
	
	// Validate signature components
	if tx.GetV() == nil || tx.GetR() == nil || tx.GetS() == nil {
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
	
	logger.Debugf("Transaction validation passed: %s", tx.GetHash())
	return nil
}

func (v *Validator) ValidateBlock(block Block) error {
	if block == nil {
		return errors.New("block is nil")
	}
	
	header := block.GetHeader()
	if header == nil {
		return errors.New("block header is nil")
	}
	
	// Validate block gas limit
	if header.GetGasLimit() > v.maxGasLimit {
		logger.Warningf("Block gas limit too high: %d", header.GetGasLimit())
		return errors.New("block gas limit too high")
	}
	
	// Validate gas used doesn't exceed limit
	if header.GetGasUsed() > header.GetGasLimit() {
		logger.Warningf("Block gas used exceeds limit: %d > %d", header.GetGasUsed(), header.GetGasLimit())
		return errors.New("block gas used exceeds limit")
	}
	
	// Validate block timestamp (should not be too far in future)
	// Allow up to 15 minutes in future
	if header.GetTimestamp() > (getCurrentTimestamp() + 900) {
		logger.Warningf("Block timestamp too far in future: %d", header.GetTimestamp())
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
	transactions := block.GetTransactions()
	for i, tx := range transactions {
		if err := v.ValidateTransaction(tx); err != nil {
			logger.Errorf("Invalid transaction %d in block: %v", i, err)
			return err
		}
		totalGasUsed += tx.GetGasLimit()
	}
	
	// Check if calculated gas matches header
	if totalGasUsed != header.GetGasUsed() {
		logger.Warningf("Block gas used mismatch: calculated %d, header %d", totalGasUsed, header.GetGasUsed())
		return errors.New("block gas used mismatch")
	}
	
	logger.Debugf("Block validation passed: %s", header.GetHash())
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
	return time.Now().Unix()
}
