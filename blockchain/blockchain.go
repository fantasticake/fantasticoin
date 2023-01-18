package blockchain

import (
	"sync"

	"github.com/fantasticake/simple-coin/db"
	"github.com/fantasticake/simple-coin/utils"
)

type blockchain struct {
	LastHash string
	m        sync.Mutex
}

type storageLayer interface {
	GetBlockchain() []byte
	SaveBlockchain(data []byte)
	ClearBlocks()
	FindBlock(key []byte) ([]byte, error)
	SaveBlock(key []byte, data []byte)
}

type dbStorage struct{}

func (dbStorage) GetBlockchain() []byte {
	return db.GetBlockchain()
}
func (dbStorage) SaveBlockchain(data []byte) {
	db.SaveBlockchain(data)
}
func (dbStorage) ClearBlocks() {
	db.ClearBlocks()
}
func (dbStorage) FindBlock(key []byte) ([]byte, error) {
	return db.FindBlock(key)
}
func (dbStorage) SaveBlock(key []byte, data []byte) {
	db.SaveBlock(key, data)
}

var (
	defaultDifficulty    int = 2
	recalcDiffInterval   int = 5 //when to recalculate difficulty per blocks
	blocksPerMin         int = 2
	blocksPerMinErrRange int = 1

	b       *blockchain
	storage storageLayer = dbStorage{}
	once    sync.Once
)

func BC() *blockchain {
	once.Do(func() {
		b = &blockchain{LastHash: ""}
		blockchainAsB := storage.GetBlockchain()
		if blockchainAsB != nil {
			utils.FromBytes(b, blockchainAsB)
		} else {
			b.AddBlock()
		}
	})
	return b
}

func PersistBlockchain(bc *blockchain) {
	storage.SaveBlockchain(utils.ToBytes(bc))
}

func isEmpty(b *blockchain) bool {
	b.m.Lock()
	defer b.m.Unlock()
	if b.LastHash == "" {
		return true
	}
	return false
}

func GetHeight(b *blockchain) int {
	if isEmpty(b) {
		return 0
	}
	return LastBlock(b).Height
}

func (b *blockchain) updateBlockchain(block *Block) {
	b.m.Lock()
	defer b.m.Unlock()
	b.LastHash = block.Hash
	PersistBlockchain(b)
}

func (b *blockchain) AddBlock() *Block {
	block := createBlock(b)
	persistBlock(block)
	b.updateBlockchain(block)
	return block
}

func (b *blockchain) AddPeerBlock(block *Block) {
	persistBlock(block)
	b.updateBlockchain(block)
	for _, tx := range block.Transactions {
		Mempool().removeTx(tx.Id)
	}
}

func Blocks(b *blockchain) []*Block {
	b.m.Lock()
	defer b.m.Unlock()
	var blocks []*Block
	hashCursor := b.LastHash
	for {
		if hashCursor == "" {
			break
		}
		block, err := FindBlock(hashCursor)
		utils.HandleErr(err)
		blocks = append(blocks, block)
		hashCursor = block.PrevHash
	}
	return blocks
}

func LastBlock(b *blockchain) *Block {
	if isEmpty(b) {
		return nil
	}
	b.m.Lock()
	defer b.m.Unlock()
	block, err := FindBlock(b.LastHash)
	utils.HandleErr(err)
	return block
}

func recalcDifficulty(b *blockchain) int {
	lastBlock := LastBlock(b)
	startBlock := Blocks(b)[recalcDiffInterval-1]
	actualTime := (lastBlock.Timestamp - startBlock.Timestamp) / 60
	aTimePerBlock := actualTime / (recalcDiffInterval - 1)
	if aTimePerBlock < blocksPerMin-blocksPerMinErrRange {
		return lastBlock.Difficulty + 1
	} else if aTimePerBlock > blocksPerMin+blocksPerMinErrRange {
		return lastBlock.Difficulty - 1
	}
	return lastBlock.Difficulty
}

func getDifficulty(b *blockchain) int {
	if isEmpty(b) {
		return defaultDifficulty
	} else if GetHeight(b)%recalcDiffInterval == 0 {
		return recalcDifficulty(b)
	} else {
		return LastBlock(b).Difficulty
	}
}

func (b *blockchain) ReplaceBlocks(blocks []*Block) {
	if len(blocks) > 0 {
		b.updateBlockchain(blocks[0])

		storage.ClearBlocks()
		for _, block := range blocks {
			persistBlock(block)
		}
	}
}
