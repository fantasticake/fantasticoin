package blockchain

import (
	"testing"

	"github.com/fantasticake/simple-coin/utils"
	"github.com/fantasticake/simple-coin/wallet"
)

type testWallet struct{}

func (testWallet) Wallet() *wallet.W {
	return &wallet.W{}
}
func (testWallet) Sign(hash string, w *wallet.W) string {
	return "signature"
}
func (testWallet) Verify(addr string, hash string, signature string) bool {
	return true
}

func TestCreateBlock(t *testing.T) {
	w = testWallet{}
	Mempool().Txs["test"] = &Tx{TxIns: []*TxIn{{TxId: "txId", Index: 0}}}
	storage = testStorage{
		fakeFindBlock: func(key []byte) ([]byte, error) {
			block := &Block{
				Transactions: []*Tx{
					{
						Id: "txId",
						TxOuts: []*TxOut{
							{},
						},
					},
				},
			}
			return utils.ToBytes(block), nil
		},
	}
	tb := createBlock(&blockchain{LastHash: "lastHash"}, 1, 1)
	if tb.PrevHash != "lastHash" {
		t.Errorf("Expected prevHash: lastHash, Got: %s", tb.PrevHash)
	}
	if tb.Height != 2 {
		t.Errorf("Expected Height: 2, Got: %d", tb.Height)
	}
	if tb.Difficulty != 1 {
		t.Errorf("Expected difficulty: 1, Got: %d", tb.Difficulty)
	}
	if len(tb.Transactions) != 2 {
		t.Errorf("Expected transaction count: 2, Got: %d", len(tb.Transactions))
	}
}
