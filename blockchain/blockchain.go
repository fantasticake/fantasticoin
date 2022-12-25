package blockchain

import (
	"sync"

	"github.com/fantasticake/fantasticoin/db"
	"github.com/fantasticake/fantasticoin/utils"
)

type blockchain struct {
	LastHash string
}

var b *blockchain
var once sync.Once

func BC() *blockchain {
	if b == nil {
		once.Do(func() {
			b = &blockchain{""}
			checkpoint := db.GetCheckpoint()
			if checkpoint != nil {
				utils.FromBytes(b, checkpoint)
			} else {
				b.AddBlock("Genesis")
			}
		})
	}
	return b
}

func PersistCheckpoint(bc *blockchain) {
	db.SaveCheckpoint(utils.ToBytes(bc))
}

func getHeight(b *blockchain) int {
	if b.LastHash == "" {
		return 0
	}
	block, err := FindBlock(b.LastHash)
	utils.HandleErr(err)
	return block.Height
}
