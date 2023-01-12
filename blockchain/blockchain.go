package blockchain

import (
	"sync"

	"github.com/fantasticake/fantasticoin/db"
	"github.com/fantasticake/fantasticoin/utils"
)

type blockchain struct {
	LastHash string
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
		b = &blockchain{""}
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

func (b *blockchain) AddBlock() {
	block := createBlock(b)
	b.LastHash = block.Hash
	persistBlock(block)
	PersistCheckpoint(b)
}

func Blocks(b *blockchain) []*Block {
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
		b.LastHash = blocks[0].Hash
		PersistCheckpoint(b)

		db.ClearBlocks()
		for _, block := range blocks {
			persistBlock(block)
		}
	}
}
