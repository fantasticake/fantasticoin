package blockchain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

type Block struct {
	Data     string `json:"data"`
	Hash     string `json:"hash"`
	PrevHash string `json:"prevHash,omitempty"`
	Height   int    `json:"height"`
}

type blockchain struct {
	blocks []*Block
}

var b *blockchain
var once sync.Once

func BC() *blockchain {
	if b == nil {
		once.Do(func() {
			b = &blockchain{}
			b.AddBlock("Genesis")
		})
	}
	return b
}

func Blocks(b *blockchain) []*Block {
	return b.blocks
}

func (b *Block) calcHash() {
	hash := sha256.Sum256([]byte(b.Data + b.PrevHash))
	b.Hash = fmt.Sprintf("%x", hash)
}

func getLatestHash(b *blockchain) string {
	totalBlocks := len(b.blocks)
	if totalBlocks == 0 {
		return ""
	}
	return b.blocks[totalBlocks-1].Hash
}

func getHeight(b *blockchain) int {
	return len(b.blocks)
}

func createBlock(b *blockchain, data string) *Block {
	newBlock := &Block{data, "", getLatestHash(b), getHeight(b) + 1}
	newBlock.calcHash()
	return newBlock
}

func (b *blockchain) AddBlock(data string) {
	b.blocks = append(b.blocks, createBlock(b, data))
}

func GetBlock(b *blockchain, height int) (*Block, error) {
	if height > getHeight(b) {
		return nil, errors.New("Block not found")
	}
	return b.blocks[height-1], nil
}
