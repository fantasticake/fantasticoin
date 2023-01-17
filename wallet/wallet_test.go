package wallet

import (
	"crypto/x509"
	"encoding/hex"
	"io/fs"
	"sync"
	"testing"
)

var (
	testWallet string = "307702010104203c13e658fc5fc6e757a307852a22b4d468b060947f9f95bfbf60e3273ddc2007a00a06082a8648ce3d030107a14403420004438ddcb272a421cf02408e9a5149c10a48b330afa63542ecc53c333558c2bff468301bb504d248b3766f19b85379fd91e334948816adce4e7281f942fd948225"
)

type testFile struct {
	FakeIsNotExist func(err error) bool
}

func (testFile) ReadFile(name string) ([]byte, error) {
	return hex.DecodeString(testWallet)
}

func (testFile) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return nil
}

func (testFile) Stat(name string) (fs.FileInfo, error) {
	return nil, nil
}

func (t testFile) IsNotExist(err error) bool {
	return t.FakeIsNotExist(err)
}

func getTestWallet() *wallet {
	keyAsB, _ := hex.DecodeString(testWallet)
	key, _ := x509.ParseECPrivateKey(keyAsB)
	tw := wallet{
		privateKey: key,
	}
	tw.calcAddr()
	return &tw
}

func TestWallet(t *testing.T) {
	t.Run("should init wallet", func(t *testing.T) {
		file = testFile{
			FakeIsNotExist: func(err error) bool { return true },
		}
		tw := Wallet()
		if tw.privateKey == nil {
			t.Errorf("privateKey should be created")
		}
		if tw.Address == "" {
			t.Errorf("address should be calculated")
		}
	})

	t.Run("should restore wallet", func(t *testing.T) {
		file = testFile{
			FakeIsNotExist: func(err error) bool { return false },
		}
		once = sync.Once{}
		tw := Wallet()
		if tw.Address != getTestWallet().Address {
			t.Errorf("should restore address, Expected: %v, Got: %v", getTestWallet().Address, tw.Address)
		}
	})
}

func TestSign(t *testing.T) {
	signature := Sign("test", getTestWallet())
	ok := Verify(getTestWallet().Address, "test", signature)
	if !ok {
		t.Error("should return correct signature")
	}
}

func TestVerify(t *testing.T) {
	signature := Sign("test", getTestWallet())
	ok := Verify(getTestWallet().Address, "test2", signature)
	if ok {
		t.Error("should return false for different data")
	}
}
