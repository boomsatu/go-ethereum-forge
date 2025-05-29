
package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
)

// StateAdapter adapts our state to vm.StateDB interface
type StateAdapter struct {
	stateDB *state.StateDB
}

func NewStateAdapter(stateDB *state.StateDB) *StateAdapter {
	return &StateAdapter{
		stateDB: stateDB,
	}
}

func (s *StateAdapter) CreateAccount(addr common.Address) {
	s.stateDB.CreateAccount(addr)
}

func (s *StateAdapter) SubBalance(addr common.Address, amount *big.Int) {
	s.stateDB.SubBalance(addr, amount)
}

func (s *StateAdapter) AddBalance(addr common.Address, amount *big.Int) {
	s.stateDB.AddBalance(addr, amount)
}

func (s *StateAdapter) GetBalance(addr common.Address) *big.Int {
	return s.stateDB.GetBalance(addr)
}

func (s *StateAdapter) GetNonce(addr common.Address) uint64 {
	return s.stateDB.GetNonce(addr)
}

func (s *StateAdapter) SetNonce(addr common.Address, nonce uint64) {
	s.stateDB.SetNonce(addr, nonce)
}

func (s *StateAdapter) GetCodeHash(addr common.Address) common.Hash {
	return s.stateDB.GetCodeHash(addr)
}

func (s *StateAdapter) GetCode(addr common.Address) []byte {
	return s.stateDB.GetCode(addr)
}

func (s *StateAdapter) SetCode(addr common.Address, code []byte) {
	s.stateDB.SetCode(addr, code)
}

func (s *StateAdapter) GetCodeSize(addr common.Address) int {
	return s.stateDB.GetCodeSize(addr)
}

func (s *StateAdapter) AddRefund(gas uint64) {
	s.stateDB.AddRefund(gas)
}

func (s *StateAdapter) SubRefund(gas uint64) {
	s.stateDB.SubRefund(gas)
}

func (s *StateAdapter) GetRefund() uint64 {
	return s.stateDB.GetRefund()
}

func (s *StateAdapter) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return s.stateDB.GetCommittedState(addr, hash)
}

func (s *StateAdapter) GetState(addr common.Address, hash common.Hash) common.Hash {
	return s.stateDB.GetState(addr, hash)
}

func (s *StateAdapter) SetState(addr common.Address, key common.Hash, value common.Hash) {
	s.stateDB.SetState(addr, key, value)
}

func (s *StateAdapter) Suicide(addr common.Address) bool {
	return s.stateDB.Suicide(addr)
}

func (s *StateAdapter) HasSuicided(addr common.Address) bool {
	return s.stateDB.HasSuicided(addr)
}

func (s *StateAdapter) Exist(addr common.Address) bool {
	return s.stateDB.Exist(addr)
}

func (s *StateAdapter) Empty(addr common.Address) bool {
	return s.stateDB.Empty(addr)
}

func (s *StateAdapter) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses []ethTypes.AccessTuple) {
	s.stateDB.PrepareAccessList(sender, dest, precompiles, txAccesses)
}

func (s *StateAdapter) AddressInAccessList(addr common.Address) bool {
	return s.stateDB.AddressInAccessList(addr)
}

func (s *StateAdapter) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return s.stateDB.SlotInAccessList(addr, slot)
}

func (s *StateAdapter) AddAddressToAccessList(addr common.Address) {
	s.stateDB.AddAddressToAccessList(addr)
}

func (s *StateAdapter) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.stateDB.AddSlotToAccessList(addr, slot)
}

func (s *StateAdapter) RevertToSnapshot(id int) {
	s.stateDB.RevertToSnapshot(id)
}

func (s *StateAdapter) Snapshot() int {
	return s.stateDB.Snapshot()
}

func (s *StateAdapter) AddLog(log *ethTypes.Log) {
	s.stateDB.AddLog(log)
}

func (s *StateAdapter) AddPreimage(hash common.Hash, preimage []byte) {
	s.stateDB.AddPreimage(hash, preimage)
}

func (s *StateAdapter) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	return s.stateDB.ForEachStorage(addr, cb)
}

// Add missing import
import ethTypes "github.com/ethereum/go-ethereum/core/types"
