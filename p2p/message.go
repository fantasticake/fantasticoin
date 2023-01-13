package p2p

import (
	"fmt"

	"github.com/fantasticake/fantasticoin/blockchain"
	"github.com/fantasticake/fantasticoin/utils"
	"github.com/gorilla/websocket"
)

const (
	lastBlockMessage = iota
	reqAllBlocksMessage
	allBlocksMessage
	newTxMessage
	newBlockMessage
	newPeerMessage
)

type message struct {
	MessageType int
	Payload     []byte
}

type NewPeerPayload struct {
	Address  string
	Port     int
	OpenPort int
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

func BroadcastNewTx(tx *blockchain.Tx) {
	Peers().m.Lock()
	defer Peers().m.Unlock()
	for _, peer := range Peers().v {
		peer.sendMessage(newTxMessage, tx)
	}
}

func BroadcastNewBlock(b *blockchain.Block) {
	Peers().m.Lock()
	defer Peers().m.Unlock()
	for _, peer := range Peers().v {
		peer.sendMessage(newBlockMessage, b)
	}
}

func BroadcastNewPeer(p *peer) {
	Peers().m.Lock()
	defer Peers().m.Unlock()
	for _, peer := range Peers().v {
		if peer != p {
			payload := &NewPeerPayload{
				Address:  p.Address,
				Port:     p.Port,
				OpenPort: peer.Port,
			}
			peer.sendMessage(newPeerMessage, payload)
		}
	}
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
	case newTxMessage:
		tx := &blockchain.Tx{}
		utils.FromJson(tx, m.Payload)
		blockchain.Mempool().Txs[tx.Id] = tx
	case newBlockMessage:
		block := &blockchain.Block{}
		utils.FromJson(block, m.Payload)
		blockchain.BC().AddPeerBlock(block)
	case newPeerMessage:
		payload := &NewPeerPayload{}
		utils.FromJson(payload, m.Payload)
		if !isConnected(payload.Address, payload.Port) {
			wsUrl := fmt.Sprintf("ws://%s:%d/ws?port=%d", payload.Address, payload.Port, payload.OpenPort)
			conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
			utils.HandleErr(err)
			Peers().InitPeer(conn, payload.Address, payload.Port)
		}
	}
}
