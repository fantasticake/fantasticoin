package blockchain

import (
	"crypto/sha256"
	"fmt"

	"github.com/fantasticake/fantasticoin/db"
	"github.com/fantasticake/fantasticoin/utils"
)

type Block struct {
	Data     string `json:"data"`
	Hash     string `json:"hash"`
	PrevHash string `json:"prevHash,omitempty"`
	Height   int    `json:"height"`
}

func (b *Block) calcHash() {
	hash := sha256.Sum256([]byte(b.Data + b.PrevHash))
	b.Hash = fmt.Sprintf("%x", hash)
}

func persistBlock(block *Block) {
	db.SaveBlock([]byte(block.Hash), utils.ToBytes(block))
}

func createBlock(b *blockchain, data string) *Block {
	newBlock := &Block{data, "", b.LastHash, getHeight(b) + 1}
	newBlock.calcHash()
	return newBlock
}

func (b *blockchain) AddBlock(data string) {
	block := createBlock(b, data)
	b.LastHash = block.Hash
	persistBlock(block)
	PersistCheckpoint(b)
}

func FindBlock(hash string) (*Block, error) {
	block := &Block{}
	hashedBlock, err := db.FindBlock([]byte(hash))
	if err != nil {
		return nil, err
	}
	utils.FromBytes(block, hashedBlock)
	return block, nil
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
