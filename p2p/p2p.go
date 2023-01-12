package p2p

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type peers struct {
	v map[string]*peer
	m sync.Mutex
}

type peer struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	conn    *websocket.Conn
	inbox   chan message
}

var (
	p    *peers
	once sync.Once
)

func Peers() *peers {
	once.Do(func() {
		p = &peers{
			v: make(map[string]*peer),
		}
	})
	return p
}

func (p *peers) InitPeer(conn *websocket.Conn, address string, port int) *peer {
	p.m.Lock()
	defer p.m.Unlock()
	newPeer := p.addPeer(conn, address, port)
	go newPeer.read()
	go newPeer.write()
	return newPeer
}

func (p *peers) addPeer(conn *websocket.Conn, address string, port int) *peer {
	newPeer := &peer{
		Address: address,
		Port:    port,
		conn:    conn,
		inbox:   make(chan message),
	}
	key := fmt.Sprintf("%s:%d", address, port)
	p.v[key] = newPeer
	return newPeer
}

func (p *peers) removePeer(address string, port int) {
	p.m.Lock()
	defer p.m.Unlock()
	key := fmt.Sprintf("%s:%d", address, port)
	p.v[key].conn.Close()
	delete(p.v, key)
}

func (p *peer) read() {
	for {
		m := message{}
		err := p.conn.ReadJSON(&m)
		if err != nil {
			Peers().removePeer(p.Address, p.Port)
			break
		}
		handleMessage(p, &m)
	}
}

func (p *peer) write() {
	defer func() {
		Peers().removePeer(p.Address, p.Port)
	}()
	for {
		payload, ok := <-p.inbox
		if !ok {
			break
		}
		err := p.conn.WriteJSON(payload)
		if err != nil {
			break
		}
	}
}

func GetPeers() []string {
	Peers().m.Lock()
	defer Peers().m.Unlock()
	var keys []string
	for key := range Peers().v {
		keys = append(keys, key)
	}
	return keys
}
