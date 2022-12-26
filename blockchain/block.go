package blockchain

import (
	"strings"
	"time"

	"github.com/fantasticake/fantasticoin/db"
	"github.com/fantasticake/fantasticoin/utils"
)

type Block struct {
	Data       string `json:"data"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prevHash,omitempty"`
	Height     int    `json:"height"`
	Difficulty int    `json:"difficulty"`
	Nonce      int    `json:"nonce"`
	Timestamp  int    `json:"timestamp"`
}

func (b *Block) mine() {
	difficulty := strings.Repeat("0", b.Difficulty)
	for {
		b.Timestamp = int(time.Now().Unix())
		b.Hash = utils.Hash(b)
		if strings.HasPrefix(b.Hash, difficulty) {
			return
		} else {
			b.Nonce += 1
		}
	}
}

func persistBlock(block *Block) {
	db.SaveBlock([]byte(block.Hash), utils.ToBytes(block))
}

func createBlock(b *blockchain, data string) *Block {
	newBlock := &Block{
		Data:       data,
		Hash:       "",
		PrevHash:   b.LastHash,
		Height:     getHeight(b) + 1,
		Difficulty: getDifficulty(b),
		Nonce:      0,
	}
	newBlock.mine()

	return newBlock
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
