// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors, Copyright 2015-2024 The go-ethereum Authors (geth)
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"github.com/holiman/uint256"
	"github.com/ixios-io/ixiosSpark/common"
)

type ContractRef interface {
	Address() common.Address
}
type AccountRef common.Address

func (ar AccountRef) Address() common.Address { return (common.Address)(ar) }

type Contract struct {
	CallerAddress common.Address
	caller        ContractRef
	self          ContractRef

	jumpdests map[common.Hash]bitvec
	analysis  bitvec

	Code     []byte
	CodeHash common.Hash
	CodeAddr *common.Address
	Input    []byte

	Gas   uint64
	value *uint256.Int
}

func NewContract(caller ContractRef, object ContractRef, value *uint256.Int, gas uint64) *Contract {
	// evm is not used
	return &Contract{}
}

func (c *Contract) validJumpdest(dest *uint256.Int) bool {
	// evm is not used
	return false
}

func (c *Contract) isCode(udest uint64) bool {
	// evm is not used
	return false
}

func (c *Contract) AsDelegate() *Contract {
	// evm is not used
	return &Contract{}
}

func (c *Contract) GetOp(n uint64) OpCode {
	// evm is not used
	return STOP
}

func (c *Contract) Caller() common.Address {
	// evm is not used
	return common.Address{}
}

// UseGas attempts the use gas and subtracts it and returns true on success
func (c *Contract) UseGas(gas uint64) (ok bool) {
	// evm is not used
	return false
}

func (c *Contract) Address() common.Address {
	return c.self.Address()
}

func (c *Contract) Value() *uint256.Int {
	return c.value
}

func (c *Contract) SetCallCode(addr *common.Address, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeAddr = addr
}

func (c *Contract) SetCodeOptionalHash(addr *common.Address, codeAndHash *codeAndHash) {
	c.Code = codeAndHash.code
	c.CodeHash = codeAndHash.hash
	c.CodeAddr = addr
}
