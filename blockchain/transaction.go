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
	Owner string `json:"owner"`
	TxId  string `json:"txId"`
	Index int    `json:"index"`
}

type TxOut struct {
	Onwer  string `json:"owner"`
	Amount int    `json:"amount"`
}

type UTxOut struct {
	TxId   string `json:"txId"`
	Index  int    `json:"index"`
	Amount int    `json:"amount"`
}

type mempool struct {
	Txs []*Tx `json:"txs"`
}

var m *mempool
var minerReward int = 10
var TestWallet string = "testWallet"

func (m *mempool) clear() {
	m.Txs = nil
}

func Mempool() *mempool {
	if m == nil {
		m = &mempool{}
	}
	return m
}

func (t *Tx) calcId() {
	t.Id = utils.Hash(t)
}

func isOnMempool(uTxOut *UTxOut) bool {
	for _, tx := range Mempool().Txs {
		for _, txIn := range tx.TxIns {
			if txIn.TxId == uTxOut.TxId && txIn.Index == uTxOut.Index {
				return true
			}
		}
	}
	return false
}

func getTxstoConfirm() []*Tx {
	txs := Mempool().Txs
	Mempool().clear()
	return append(txs, makeCoinbaseTx())
}

func makeCoinbaseTx() *Tx {
	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns: []*TxIn{{
			Owner: "Coinbase",
			TxId:  "",
			Index: -1,
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
	var total int
	uTxOuts := GetUTxOutsByAddr(b, TestWallet)
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := TxIn{
			Owner: TestWallet,
			TxId:  uTxOut.TxId,
			Index: uTxOut.Index,
		}
		txIns = append(txIns, &txIn)
		total += uTxOut.Amount
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

func GetUTxOutsByAddr(b *blockchain, address string) []*UTxOut {
	var uTxOuts []*UTxOut
	txUsedMap := make(map[string]bool)
	for _, block := range Blocks(BC()) {
		for _, tx := range block.Transactions {
			for _, txIn := range tx.TxIns {
				if txIn.Owner == address {
					txUsedMap[txIn.TxId] = true
				}
			}
			for index, txOut := range tx.TxOuts {
				if txOut.Onwer == address {
					if _, ok := txUsedMap[tx.Id]; !ok {
						uTxOut := &UTxOut{
							TxId:   tx.Id,
							Index:  index,
							Amount: txOut.Amount,
						}
						if !isOnMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

func GetBalanceByAddr(b *blockchain, address string) int {
	uTxOuts := GetUTxOutsByAddr(b, address)
	var total int
	for _, txOut := range uTxOuts {
		total += txOut.Amount
	}
	return total
}
