package blockchain

import (
	"sync"
	"testing"

	"github.com/fantasticake/simple-coin/utils"
)

type testStorage struct {
	fakeGetBlockchain func() []byte
	fakeFindBlock     func(key []byte) ([]byte, error)
}

func (t testStorage) GetBlockchain() []byte {
	return t.fakeGetBlockchain()
}
func (testStorage) SaveBlockchain(data []byte) {}
func (testStorage) ClearBlocks()               {}
func (t testStorage) FindBlock(key []byte) ([]byte, error) {
	return t.fakeFindBlock(key)
}
func (testStorage) SaveBlock(key []byte, data []byte) {}

func TestBC(t *testing.T) {
	t.Run("should return a new blockchain with a block", func(t *testing.T) {
		storage = testStorage{
			fakeGetBlockchain: func() []byte { return nil },
		}
		tb := BC()
		if tb.LastHash == "" {
			t.Error("should have a lastHash")
		}
	})

	t.Run("should return a restored blockchain", func(t *testing.T) {
		once = sync.Once{}
		storage = testStorage{
			fakeGetBlockchain: func() []byte {
				return utils.ToBytes(blockchain{LastHash: "test"})
			},
		}
		tb := BC()
		if tb.LastHash != "test" {
			t.Errorf("lastHash should be restored, Expected: test, Got: %s", tb.LastHash)
		}
	})
}

func TestGetDifficulty(t *testing.T) {
	t.Run("should reuturn a defaultDifficulty if blockchain is empty", func(t *testing.T) {
		d := getDifficulty(&blockchain{LastHash: ""})
		if d != defaultDifficulty {
			t.Errorf("Expected: %d, Got: %d", defaultDifficulty, d)
		}
	})

	t.Run("should recalculate difficulty", func(t *testing.T) {
		blockCur := 0
		storage = testStorage{
			fakeFindBlock: func(key []byte) ([]byte, error) {
				blocks := []*Block{
					{Height: recalcDiffInterval},
					{Difficulty: 1},
					{PrevHash: "prevHash"},
					{PrevHash: "prevHash"},
					{PrevHash: "prevHash"},
					{PrevHash: "prevHash"},
					{},
				}
				blockAsB := utils.ToBytes(blocks[blockCur])
				blockCur += 1
				return blockAsB, nil
			},
		}
		d := getDifficulty(&blockchain{LastHash: "test"})
		if d != 2 {
			t.Errorf("Expected: 2, Got: %d", d)
		}
	})
	t.Run("should return a last blocks's difficulty", func(t *testing.T) {
		blockCur := 0
		storage = testStorage{
			fakeFindBlock: func(key []byte) ([]byte, error) {
				blocks := []*Block{
					{Height: recalcDiffInterval + 1},
					{Difficulty: 1},
				}
				block := blocks[blockCur]
				blockCur += 1
				return utils.ToBytes(block), nil
			},
		}
		d := getDifficulty(&blockchain{LastHash: "lastHash"})
		if d != 1 {
			t.Errorf("Expected: 1, Got: %d", d)
		}
	})
}
