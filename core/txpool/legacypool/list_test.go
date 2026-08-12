package legacypool

import (
	"math/big"
	"testing"

	"github.com/ixios-io/ixiosSpark/common"
	"github.com/ixios-io/ixiosSpark/core/types"
)

func newTestLegacyTx(nonce uint64, gasPrice int64) *types.Transaction {
	to := common.HexToAddress("0x1")
	return types.NewTransaction(nonce, to, big.NewInt(0), 21000, big.NewInt(gasPrice), nil)
}

func TestListFilterGasPriceStrict(t *testing.T) {
	list := newList(true)
	for _, tx := range []*types.Transaction{
		newTestLegacyTx(0, 99),
		newTestLegacyTx(1, 100),
		newTestLegacyTx(2, 101),
	} {
		if inserted, _ := list.Add(tx, 10); !inserted {
			t.Fatalf("failed to add tx nonce %d", tx.Nonce())
		}
	}

	removed, invalids := list.FilterGasPrice(big.NewInt(100))
	if len(removed) != 1 || removed[0].Nonce() != 0 {
		t.Fatalf("unexpected removed txs: %+v", removed)
	}
	if len(invalids) != 2 || invalids[0].Nonce() != 1 || invalids[1].Nonce() != 2 {
		t.Fatalf("unexpected invalidated txs: %+v", invalids)
	}
	if !list.Empty() {
		t.Fatalf("strict list should be empty after invalidating the suffix")
	}
}

func TestListFilterGasPriceQueue(t *testing.T) {
	list := newList(false)
	for _, tx := range []*types.Transaction{
		newTestLegacyTx(0, 99),
		newTestLegacyTx(1, 100),
		newTestLegacyTx(2, 101),
	} {
		if inserted, _ := list.Add(tx, 10); !inserted {
			t.Fatalf("failed to add tx nonce %d", tx.Nonce())
		}
	}

	removed, invalids := list.FilterGasPrice(big.NewInt(100))
	if len(removed) != 1 || removed[0].Nonce() != 0 {
		t.Fatalf("unexpected removed txs: %+v", removed)
	}
	if len(invalids) != 0 {
		t.Fatalf("queue list should not invalidate suffix txs: %+v", invalids)
	}
	remaining := list.Flatten()
	if len(remaining) != 2 || remaining[0].Nonce() != 1 || remaining[1].Nonce() != 2 {
		t.Fatalf("unexpected remaining txs: %+v", remaining)
	}
}
