package p2p

import (
	"github.com/fantasticake/fantasticoin/blockchain"
	"github.com/fantasticake/fantasticoin/utils"
)

const (
	lastBlockMessage = iota
	reqAllBlocksMessage
	allBlocksMessage
)

type message struct {
	MessageType int
	Payload     []byte
}

func (p *peer) sendMessage(messageType int, payload any) {
	m := message{messageType, utils.ToJson(payload)}
	p.inbox <- m
}

func (p *peer) SendLastBlock() {
	p.sendMessage(lastBlockMessage, blockchain.LastBlock(blockchain.BC()))
}

func (p *peer) requestAllBlocks() {
	p.sendMessage(reqAllBlocksMessage, nil)
}

func (p *peer) sendAllBlocks() {
	p.sendMessage(allBlocksMessage, blockchain.Blocks(blockchain.BC()))
}

func handleMessage(p *peer, m *message) {
	switch m.MessageType {
	case lastBlockMessage:
		block := blockchain.Block{}
		utils.FromJson(&block, m.Payload)
		if block.Height >= blockchain.GetHeight(blockchain.BC()) {
			p.requestAllBlocks()
		} else {
			p.sendAllBlocks()
		}
	case reqAllBlocksMessage:
		p.sendAllBlocks()
	case allBlocksMessage:
		blocks := []*blockchain.Block{}
		utils.FromJson(&blocks, m.Payload)
		if len(blocks) >= blockchain.GetHeight(blockchain.BC()) {
			blockchain.BC().ReplaceBlocks(blocks)
		}
	}
}
