package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fantasticake/fantasticoin/blockchain"
	"github.com/fantasticake/fantasticoin/utils"
	"github.com/gorilla/mux"
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

type blocksInput struct {
	Data string
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
			Payload:     "data:string",
		},
		{
			Url:         URL("/blocks/{height}"),
			Method:      "GET",
			Description: "get a block by height",
		},
	}

	utils.HandleErr(json.NewEncoder(w).Encode(urlData))
}

func blocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		blocks := blockchain.Blocks(blockchain.BC())
		utils.HandleErr(json.NewEncoder(w).Encode(blocks))
	case "POST":
		input := blocksInput{}
		utils.HandleErr(json.NewDecoder(r.Body).Decode(&input))
		blockchain.BC().AddBlock(input.Data)
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
	router.HandleFunc("/blocks", blocks).Methods("GET", "POST")
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET")

	fmt.Printf("Server listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
