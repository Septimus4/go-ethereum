// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eip1559

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// VerifyEIP1559Header verifies some header attributes which were changed in EIP-1559,
// - gas limit check
// - basefee check
func VerifyEIP1559Header(config *params.ChainConfig, parent, header *types.Header) error {
	// Verify that the gas limit remains within allowed bounds
	parentGasLimit := parent.GasLimit
	if !config.IsLondon(parent.Number) {
		parentGasLimit = parent.GasLimit * config.ElasticityMultiplier()
	}
	if err := misc.VerifyGaslimit(parentGasLimit, header.GasLimit); err != nil {
		return err
	}
	// Verify the header is not malformed
	if header.BaseFee == nil {
		return errors.New("header is missing baseFee")
	}
	// Verify the baseFee is correct based on the parent header.
	expectedBaseFee := CalcBaseFee(config, parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			header.BaseFee, expectedBaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

// CalcBaseFee calculates the basefee of the header.
func CalcBaseFee(config *params.ChainConfig, parent *types.Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsLondon(parent.Number) {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}

	parentGasTarget := config.GasTarget(parent.GasLimit)
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}

	// Compute the full denominator as (parentGasTarget * BaseFeeChangeDenominator)
	denom := new(big.Int).Mul(
		new(big.Int).SetUint64(parentGasTarget),
		new(big.Int).SetUint64(config.BaseFeeChangeDenominator()),
	)

	delta := new(big.Int)
	// Increase: use a slope of 3/8.
	if parent.GasUsed > parentGasTarget {
		diff := parent.GasUsed - parentGasTarget
		// delta = parent.BaseFee * diff * 3 / (parentGasTarget * BaseFeeChangeDenominator)
		delta.SetUint64(diff)
		delta.Mul(delta, parent.BaseFee)
		delta.Mul(delta, big.NewInt(3))
		delta.Div(delta, denom)
		// Enforce a minimum increase of 1.
		if delta.Cmp(common.Big1) < 0 {
			delta = common.Big1
		}
		return new(big.Int).Add(parent.BaseFee, delta)
	}

	// Decrease: use a slope of 1/8.
	// delta = parent.BaseFee * (parentGasTarget - parent.GasUsed) / (parentGasTarget * BaseFeeChangeDenominator)
	diff := parentGasTarget - parent.GasUsed
	delta.SetUint64(diff)
	delta.Mul(delta, parent.BaseFee)
	delta.Div(delta, denom)

	baseFee := delta.Sub(parent.BaseFee, delta)
	if baseFee.Cmp(common.Big0) < 0 {
		baseFee = common.Big0
	}
	return baseFee
}
