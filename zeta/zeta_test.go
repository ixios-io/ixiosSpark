package zeta

import (
	"math/big"
	"testing"

	"github.com/ixios-io/ixiosSpark/params"
)

func TestMinimumGasPriceZeta(t *testing.T) {
	preAegisBlock := new(big.Int).Sub(new(big.Int).Set(params.MainnetChainConfig.AegisBlock), big.NewInt(1))
	if got := MinimumGasPriceZeta(params.MainnetChainConfig, preAegisBlock); got != preAegisMinimumGasPriceZeta {
		t.Fatalf("pre-Aegis multiplier mismatch: got %d want %d", got, preAegisMinimumGasPriceZeta)
	}
	if got := MinimumGasPriceZeta(params.MainnetChainConfig, params.MainnetChainConfig.AegisBlock); got != postAegisMinimumGasPriceZeta {
		t.Fatalf("Aegis multiplier mismatch: got %d want %d", got, postAegisMinimumGasPriceZeta)
	}
}

func TestCalculateMinimumGasPrice(t *testing.T) {
	block := new(big.Int).Set(params.MainnetChainConfig.AegisBlock)
	oneZeta := CalculateZetaValue(block)
	want := new(big.Int).Mul(new(big.Int).Set(oneZeta), big.NewInt(postAegisMinimumGasPriceZeta))
	got := CalculateMinimumGasPrice(params.MainnetChainConfig, block)
	if got.Cmp(want) != 0 {
		t.Fatalf("minimum gas price mismatch: got %v want %v", got, want)
	}
}
