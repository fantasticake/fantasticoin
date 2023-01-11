package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fantasticake/fantasticoin/blockchain"
	"github.com/fantasticake/fantasticoin/p2p"
	"github.com/fantasticake/fantasticoin/utils"
	"github.com/fantasticake/fantasticoin/wallet"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var port int

type URL string

func (u URL) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("http://localhost:%d%s", port, u)), nil
}

type urlInfo struct {
	Url         URL    `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

type totalBalanceResponse struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type sendPayload struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type connectPayload struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func documentaion(w http.ResponseWriter, r *http.Request) {
	urlData := []urlInfo{
		{
			Url:         URL("/blocks"),
			Method:      "GET",
			Description: "Get all blocks",
		},
		{
			Url:         URL("/blocks"),
			Method:      "POST",
			Description: "Add a block",
		},
		{
			Url:         URL("/blocks/{hash}"),
			Method:      "GET",
			Description: "get a block by hash",
		},
	}

	utils.HandleErr(json.NewEncoder(w).Encode(urlData))
}

func balance(w http.ResponseWriter, r *http.Request) {
	isTotal := r.URL.Query().Get("total")
	encoder := json.NewEncoder(w)
	switch isTotal {
	case "true":
		utils.HandleErr(encoder.Encode(totalBalanceResponse{
			Address: wallet.Wallet().Address,
			Amount:  blockchain.GetBalanceByAddr(blockchain.BC(), wallet.Wallet().Address),
		}))
	default:
		utils.HandleErr(encoder.Encode(blockchain.GetUTxOutsByAddr(blockchain.BC(), wallet.Wallet().Address)))
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	var payload sendPayload
	utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload))
	err := blockchain.Mempool().AddTx(blockchain.BC(), payload.To, payload.Amount)
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{fmt.Sprint(err)})
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func mempool(w http.ResponseWriter, r *http.Request) {
	utils.HandleErr(json.NewEncoder(w).Encode(blockchain.Mempool()))
}

func blocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		blocks := blockchain.Blocks(blockchain.BC())
		utils.HandleErr(json.NewEncoder(w).Encode(blocks))
	case "POST":
		blockchain.BC().AddBlock()
		w.WriteHeader(http.StatusCreated)
	}
}

func block(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	block, err := blockchain.FindBlock(hash)
	encoder := json.NewEncoder(w)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		utils.HandleErr(encoder.Encode(errorResponse{fmt.Sprint(err)}))
	} else {
		utils.HandleErr(encoder.Encode(block))
	}
}

func peers(w http.ResponseWriter, r *http.Request) {
	utils.HandleErr(json.NewEncoder(w).Encode(p2p.Peers))
}

func ws(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	utils.HandleErr(err)
	address := strings.Split(r.RemoteAddr, ":")[0]
	port, err := strconv.Atoi(r.URL.Query().Get("port"))
	utils.HandleErr(err)
	p2p.AddPeer(conn, address, port)
}

func connect(w http.ResponseWriter, r *http.Request) {
	payload := connectPayload{}
	utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload))
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%d/ws?port=%d", payload.Address, payload.Port, port), nil)
	utils.HandleErr(err)
	p2p.AddPeer(conn, payload.Address, payload.Port)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Start(aPort int) {
	port = aPort
	router := mux.NewRouter()
	router.Use(jsonMiddleware)
	router.HandleFunc("/", documentaion).Methods("GET")
	router.HandleFunc("/balance", balance).Methods("GET")
	router.HandleFunc("/send", send).Methods("POST")
	router.HandleFunc("/mempool", mempool).Methods("GET")
	router.HandleFunc("/blocks", blocks).Methods("GET", "POST")
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET")
	router.HandleFunc("/peers", peers).Methods("GET")
	router.HandleFunc("/ws", ws).Methods("GET")
	router.HandleFunc("/connect", connect).Methods("POST")

	fmt.Printf("Server listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
