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

var (
	defaultDifficulty    int = 2
	recalcDiffInterval   int = 5 //when to recalculate difficulty per blocks
	blocksPerMin         int = 2
	blocksPerMinErrRange int = 1

	b    *blockchain
	once sync.Once
)

func BC() *blockchain {
	once.Do(func() {
		b = &blockchain{LastHash: ""}
		checkpoint := db.GetCheckpoint()
		if checkpoint != nil {
			utils.FromBytes(b, checkpoint)
		} else {
			b.AddBlock()
		}
	})
	return b
}

func isEmpty(b *blockchain) bool {
	b.m.Lock()
	defer b.m.Unlock()
	if b.LastHash == "" {
		return true
	}
	return false
}

func PersistCheckpoint(bc *blockchain) {
	db.SaveCheckpoint(utils.ToBytes(bc))
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
	PersistCheckpoint(b)
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

		db.ClearBlocks()
		for _, block := range blocks {
			persistBlock(block)
		}
	}
}
