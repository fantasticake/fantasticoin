package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/fantasticake/simple-coin/utils"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

var (
	walletFile string         = "simple_coin.wallet"
	ec         elliptic.Curve = elliptic.P256()
	w          *wallet
	once       sync.Once
)

func Wallet() *wallet {
	once.Do(func() {
		w = &wallet{}
		if fileExists(walletFile) {
			w.restore()
		} else {
			w.init()
		}
	})
	return w
}

func Sign(hash string) string {
	r, s, err := ecdsa.Sign(rand.Reader, Wallet().privateKey, utils.ToBytes(hash))
	utils.HandleErr(err)
	return fmt.Sprintf("%x", append(r.Bytes(), s.Bytes()...))
}

func Verify(addr, hash, signature string) bool {
	x, y := bigIntsByHexStr(addr)
	pub := ecdsa.PublicKey{
		Curve: ec,
		X:     x,
		Y:     y,
	}
	r, s := bigIntsByHexStr(signature)
	return ecdsa.Verify(&pub, utils.ToBytes(hash), r, s)
}

func bigIntsByHexStr(data string) (*big.Int, *big.Int) {
	dataAsB, err := hex.DecodeString(data)
	utils.HandleErr(err)
	x := big.Int{}
	y := big.Int{}
	x.SetBytes(dataAsB[:len(dataAsB)/2])
	y.SetBytes(dataAsB[len(dataAsB)/2:])
	return &x, &y
}

func (w *wallet) restore() {
	w.restoreKey()
	w.calcAddr()
}

func (w *wallet) calcAddr() {
	xAsB := w.privateKey.X.Bytes()
	yAsB := w.privateKey.Y.Bytes()
	w.Address = fmt.Sprintf("%x", append(xAsB, yAsB...))
}

func (w *wallet) init() {
	w.createKey()
	persistKey(w.privateKey)
	w.calcAddr()
}

func (w *wallet) createKey() {
	key, err := ecdsa.GenerateKey(ec, rand.Reader)
	utils.HandleErr(err)
	w.privateKey = key
}

func (w *wallet) restoreKey() {
	keyAsB, err := os.ReadFile(walletFile)
	utils.HandleErr(err)
	key, err := x509.ParseECPrivateKey(keyAsB)
	utils.HandleErr(err)
	w.privateKey = key
}

func persistKey(key *ecdsa.PrivateKey) {
	keyAsB, err := x509.MarshalECPrivateKey(key)
	utils.HandleErr(err)
	utils.HandleErr(os.WriteFile(walletFile, keyAsB, 0700))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		utils.HandleErr(err)
	}
	return true
}
