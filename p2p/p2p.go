package p2p

import "github.com/gorilla/websocket"

type peer struct {
	conn    *websocket.Conn
	Address string `json:"address"`
	Port    int    `json:"port"`
}

var Peers []*peer

func AddPeer(conn *websocket.Conn, address string, port int) {
	p := &peer{conn, address, port}
	Peers = append(Peers, p)
}
