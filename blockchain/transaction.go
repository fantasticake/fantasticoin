package blockchain

import (
	"errors"
	"time"

	"github.com/fantasticake/fantasticoin/utils"
	"github.com/fantasticake/fantasticoin/wallet"
)

type Tx struct {
	Id        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOuts"`
}

type TxIn struct {
	Address   string `json:"address"`
	TxId      string `json:"txId"`
	Index     int    `json:"index"`
	Signature string `json:"signature,omitempty"`
}

type TxOut struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
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

func getTxstoConfirm(b *blockchain) []*Tx {
	var txs []*Tx
	for _, tx := range Mempool().Txs {
		if verifyTx(b, tx) {
			txs = append(txs, tx)
		}
	}
	Mempool().clear()
	return append(txs, makeCoinbaseTx())
}

func makeCoinbaseTx() *Tx {
	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns: []*TxIn{{
			Address: "Coinbase",
			TxId:    "",
			Index:   -1,
		}},
		TxOuts: []*TxOut{{
			Address: wallet.Wallet().Address,
			Amount:  minerReward,
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
	if GetBalanceByAddr(b, wallet.Wallet().Address) < amount {
		return nil, errors.New("Not enough balance")
	}

	txIns := []*TxIn{}
	var total int
	uTxOuts := GetUTxOutsByAddr(b, wallet.Wallet().Address)
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := TxIn{
			Address: wallet.Wallet().Address,
			TxId:    uTxOut.TxId,
			Index:   uTxOut.Index,
		}
		txIns = append(txIns, &txIn)
		total += uTxOut.Amount
	}

	txOuts := []*TxOut{}
	change := total - amount
	if change > 0 {
		txOut := TxOut{
			Address: wallet.Wallet().Address,
			Amount:  change,
		}
		txOuts = append(txOuts, &txOut)
	}
	txOut := TxOut{
		Address: to,
		Amount:  amount,
	}
	txOuts = append(txOuts, &txOut)

	tx := Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.calcId()
	tx.sign()
	return &tx, nil
}

func findTx(b *blockchain, id string) *Tx {
	blocks := Blocks(b)
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if tx.Id == id {
				return tx
			}
		}
	}
	return nil
}

func (t *Tx) sign() {
	for _, txIn := range t.TxIns {
		txIn.Signature = wallet.Sign(t.Id)
	}
}

func verifyTx(b *blockchain, t *Tx) bool {
	for _, txIn := range t.TxIns {
		txOut := findTx(b, txIn.TxId).TxOuts[txIn.Index]
		ok := wallet.Verify(txOut.Address, t.Id, txIn.Signature)
		if !ok {
			return false
		}
	}
	return true
}

func GetUTxOutsByAddr(b *blockchain, address string) []*UTxOut {
	var uTxOuts []*UTxOut
	txUsedMap := make(map[string]bool)
	for _, block := range Blocks(BC()) {
		for _, tx := range block.Transactions {
			for _, txIn := range tx.TxIns {
				if txIn.Address == address {
					txUsedMap[txIn.TxId] = true
				}
			}
			for index, txOut := range tx.TxOuts {
				if txOut.Address == address {
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
