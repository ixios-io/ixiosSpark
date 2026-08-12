// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
// Copyright 2025 The ixiosSpark Authors
// You should have received a copy of the GNU Affero General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package zeta

import (
	"math/big"

	"github.com/ixios-io/ixiosSpark/params"
)

const (
	secondsPerYear = 31536000
	decayRate      = 930000000000000000

	preAegisMinimumGasPriceZeta  int64 = 1
	postAegisMinimumGasPriceZeta int64 = 100
)

// MinimumGasPriceZeta returns the required minimum gas price in zeta
// units for the given block. Before Aegis the floor is 1 zeta. Once Aegis is
// active, the floor becomes 100 zeta.
func MinimumGasPriceZeta(config *params.ChainConfig, blockNumber *big.Int) int64 {
	if config != nil && blockNumber != nil && config.IsAddressFormat(blockNumber) {
		return postAegisMinimumGasPriceZeta
	}
	return preAegisMinimumGasPriceZeta
}

// CalculateMinimumGasPrice returns the minimum transaction gas price in wei for
// the given block, applying the Aegis fork multiplier to the underlying zeta
// value.
func CalculateMinimumGasPrice(config *params.ChainConfig, blockNumber *big.Int) *big.Int {
	minimumGasPrice := CalculateZetaValue(blockNumber)
	multiplier := MinimumGasPriceZeta(config, blockNumber)
	if multiplier == 1 {
		return minimumGasPrice
	}
	return minimumGasPrice.Mul(minimumGasPrice, big.NewInt(multiplier))
}

func CalculateZetaValue(blockNumber *big.Int) *big.Int {
	// Constants
	baseValue, _ := new(big.Int).SetString("1000000000000000", 10)
	minValue := new(big.Int).SetUint64(1)
	blocksPerYear := new(big.Int).SetUint64(secondsPerYear) // Assumes one block per second

	// Calculate the 0-based year index
	yearIndex := new(big.Int).Div(blockNumber, blocksPerYear)

	// If we're in the first year, return the base value.
	if yearIndex.Sign() == 0 {
		return new(big.Int).Set(baseValue)
	}

	// Calculate the difference between base and minimum values.
	valueDiff := new(big.Int).Sub(baseValue, minValue)

	// Fixed-point arithmetic parameters (using 18 decimals of precision).
	precision := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	decayRate := new(big.Int).SetUint64(decayRate)

	// Calculate (decayRate ^ yearIndex) using fixed-point arithmetic.
	// Start with 1.0 in fixed-point representation.
	rate := new(big.Int).Set(precision)
	one := big.NewInt(1)
	for i := new(big.Int); i.Cmp(yearIndex) < 0; i.Add(i, one) {
		rate.Mul(rate, decayRate)
		rate.Div(rate, precision)
	}

	// Multiply the value difference by the computed rate, then adjust for precision.
	valueDiff.Mul(valueDiff, rate)
	valueDiff.Div(valueDiff, precision)

	// Calculate the final result: minValue + (valueDiff)
	result := new(big.Int).Add(minValue, valueDiff)
	return result
}
