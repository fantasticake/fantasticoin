package blockchain

import (
	"errors"
	"time"

	"github.com/fantasticake/fantasticoin/utils"
)

type Tx struct {
	Id        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOuts"`
}

type TxIn struct {
	Owner  string `json:"owner"`
	Amount int    `json:"amount"`
}

type TxOut struct {
	Onwer  string `json:"owner"`
	Amount int    `json:"amount"`
}

type mempool struct {
	Txs []*Tx `json:"txs"`
}

var m *mempool
var minerReward int = 10
var TestWallet string = "testWallet"

func Mempool() *mempool {
	if m == nil {
		m = &mempool{}
	}
	return m
}

func (t *Tx) calcId() {
	t.Id = utils.Hash(t)
}

func makeCoinbaseTx() *Tx {
	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns: []*TxIn{{
			Owner:  "Coinbase",
			Amount: minerReward,
		}},
		TxOuts: []*TxOut{{
			Onwer:  TestWallet,
			Amount: minerReward,
		}},
	}
	tx.calcId()

	return tx
}

func (m *mempool) AddTx(b *blockchain, to string, amount int) error {
	tx, err := makeTx(b, to, amount)
	if err != nil {
		return err
	}
	m.Txs = append(m.Txs, tx)
	return nil
}

func makeTx(b *blockchain, to string, amount int) (*Tx, error) {
	if GetBalanceByAddr(b, TestWallet) < amount {
		return nil, errors.New("Not enough balance")
	}
	txIns := []*TxIn{}
	ownedTxOuts := GetTxOutsByAddr(b, TestWallet)
	var total int
	for _, txOut := range ownedTxOuts {
		if total >= amount {
			break
		}
		txIn := TxIn{
			Owner:  txOut.Onwer,
			Amount: txOut.Amount,
		}
		txIns = append(txIns, &txIn)
		total += txOut.Amount
	}

	txOuts := []*TxOut{}
	change := total - amount
	if change > 0 {
		txOut := TxOut{
			Onwer:  TestWallet,
			Amount: change,
		}
		txOuts = append(txOuts, &txOut)
	}
	txOut := TxOut{
		Onwer:  to,
		Amount: amount,
	}
	txOuts = append(txOuts, &txOut)

	tx := Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.calcId()

	return &tx, nil
}

func getTxOuts(b *blockchain) []*TxOut {
	var txOuts []*TxOut
	blocks := Blocks(b)
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			txOuts = append(txOuts, tx.TxOuts...)
		}
	}
	return txOuts
}

func GetTxOutsByAddr(b *blockchain, address string) []*TxOut {
	var ownedTxOuts []*TxOut
	txOuts := getTxOuts(b)
	for _, txOut := range txOuts {
		if txOut.Onwer == address {
			ownedTxOuts = append(ownedTxOuts, txOut)
		}
	}
	return ownedTxOuts
}

func GetBalanceByAddr(b *blockchain, address string) int {
	ownedTxOuts := GetTxOutsByAddr(b, address)
	var total int
	for _, txOut := range ownedTxOuts {
		total += txOut.Amount
	}
	return total
}
