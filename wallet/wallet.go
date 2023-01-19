package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"sync"

	"github.com/fantasticake/simple-coin/utils"
)

type fileLayer interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	IsNotExist(err error) bool
}

type osFile struct{}

func (osFile) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osFile) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osFile) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (osFile) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type W struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

var (
	file       fileLayer      = osFile{}
	walletFile string         = "simple_coin.wallet"
	ec         elliptic.Curve = elliptic.P256()
	w          *W
	once       sync.Once
)

func Wallet() *W {
	once.Do(func() {
		w = &W{}
		if fileExists(walletFile) {
			w.restore()
		} else {
			w.init()
		}
	})
	return w
}

func Sign(hash string, w *W) string {
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, utils.ToBytes(hash))
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

func (w *W) restore() {
	w.restoreKey()
	w.calcAddr()
}

func (w *W) calcAddr() {
	xAsB := w.privateKey.X.Bytes()
	yAsB := w.privateKey.Y.Bytes()
	w.Address = fmt.Sprintf("%x", append(xAsB, yAsB...))
}

func (w *W) init() {
	w.createKey()
	persistKey(w.privateKey)
	w.calcAddr()
}

func (w *W) createKey() {
	key, err := ecdsa.GenerateKey(ec, rand.Reader)
	utils.HandleErr(err)
	w.privateKey = key
}

func (w *W) restoreKey() {
	keyAsB, err := file.ReadFile(walletFile)
	utils.HandleErr(err)
	key, err := x509.ParseECPrivateKey(keyAsB)
	utils.HandleErr(err)
	w.privateKey = key
}

func persistKey(key *ecdsa.PrivateKey) {
	keyAsB, err := x509.MarshalECPrivateKey(key)
	utils.HandleErr(err)
	utils.HandleErr(file.WriteFile(walletFile, keyAsB, 0700))
}

func fileExists(filename string) bool {
	_, err := file.Stat(filename)
	if file.IsNotExist(err) {
		return false
	} else if err != nil {
		utils.HandleErr(err)
	}
	return true
}
